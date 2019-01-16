package server

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	_RedirectStatusReady RedirectStatus = iota
	_RedirectStatusRunning
	_RedirectStatusStoping
)

// RedirectStatus Redirect status
type RedirectStatus byte

// Redirect redirect with buffer
type Redirect struct {
	retrytimes  int
	status      RedirectStatus
	statusMu    sync.Mutex
	parallel    int
	buffer      *Buffer
	requestPool *sync.Pool
	stopChan    chan int
}

// NewRedirect redirect new
func NewRedirect(retrytimes int, parallel int, bufferLen int) *Redirect {
	r := &Redirect{
		retrytimes:  retrytimes,
		parallel:    parallel,
		status:      _RedirectStatusReady,
		buffer:      NewBuffer(bufferLen),
		requestPool: new(sync.Pool),
		stopChan:    make(chan int, parallel),
	}
	r.requestPool.New = func() interface{} {
		return &RequestWrapper{}
	}
	return r
}

func (r *Redirect) reply(conn net.Conn, req *http.Request, id string, err error) {

	respStr := "HTTP/1.1 "

	if req != nil {
		respStr = req.Proto + " "
	}

	bodyM := map[string]interface{}{
		"code":    200,
		"message": "OK",
		"data": map[string]interface{}{
			"id": id,
		},
	}

	if err != nil {
		bodyM["code"] = 500
		bodyM["message"] = err.Error()
		bodyM["data"] = nil
		respStr += "500 Internal Server Error"
	} else {
		respStr += "200 OK"
	}

	bodyB, _ := json.Marshal(bodyM)

	respStr += "\r\n"
	respStr += fmt.Sprintf("Content-Length: %d\r\n", len(bodyB))
	respStr += "Content-Type: application/json; charset=utf-8\r\n"
	respStr += "\r\n"
	respStr += string(bodyB)

	conn.Write([]byte(respStr))

}

func (r *Redirect) makeID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X-%X", uint32(time.Now().UTC().Unix()), b[0:2], b[2:4], b[4:6], b[6:8], b[8:])
}

func (r *Redirect) putTask(rw *RequestWrapper, must bool) bool {

	if must {
		for {
			if r.buffer.Put(rw, time.Now().Add(time.Second)) {
				r.logRequestWrapper("put", rw)
				return true
			}
			time.Sleep(time.Millisecond)
		}
	} else {
		if r.buffer.Put(rw, time.Now().Add(time.Second)) {
			r.logRequestWrapper("put", rw)
			return true
		}
	}
	return false
}

func (r *Redirect) getTask() *RequestWrapper {
	obj, ok := r.buffer.Get(time.Now().Add(time.Second))
	if ok {
		return obj.(*RequestWrapper)
	}
	return nil
}

// Run run task
func (r *Redirect) Run() error {
	r.statusMu.Lock()
	defer r.statusMu.Unlock()
	switch r.status {
	case _RedirectStatusReady:
		for i := 0; i < r.parallel; i++ {
			go r.do(i)
		}
	case _RedirectStatusRunning, _RedirectStatusStoping:
		return errors.New("Redirect is running or stoping")
	}
	return nil
}

// Stop stop task
func (r *Redirect) Stop() {
	r.statusMu.Lock()
	defer func() {
		r.status = _RedirectStatusReady
		r.statusMu.Unlock()

	}()
	if r.status == _RedirectStatusRunning {
		r.status = _RedirectStatusStoping
		for i := 0; i < r.parallel; i++ {
			<-r.stopChan
		}
	}

}

// Forward forward http request
func (r *Redirect) Forward(conn net.Conn) {
	if conn == nil {
		return
	}
	defer conn.Close()

	var (
		request    = []byte{}
		requestTmp = make([]byte, 1024)
	)

	for {
		n, err := conn.Read(requestTmp)
		if err != nil {
			r.reply(conn, nil, "", err)
			return
		}
		request = append(request, requestTmp[:n]...)
		if n < 1024 {
			break
		}
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(request)))

	if err != nil {
		r.reply(conn, nil, "", err)
		return
	}

	u, err := url.Parse(string(req.RequestURI[1:]))
	if err != nil {
		r.reply(conn, req, "", err)
		return
	}
	req.URL = u
	req.Host = u.Host

	req.RequestURI = ""
	requestObj := r.requestPool.Get().(*RequestWrapper)
	requestObj.CreatedTime = time.Now()
	requestObj.request = req
	requestObj.TryCount = 0
	requestObj.ID = r.makeID()

	if !r.putTask(requestObj, false) {
		r.reply(conn, req, "", errors.New("request put into buffer timeout"))
		return
	}

	r.reply(conn, req, requestObj.ID, nil)
}

func (r *Redirect) do(i int) {

	defer func() { r.stopChan <- i }()
	var (
		rw *RequestWrapper
	)
	for {
		rw = r.getTask()
		if rw == nil {
			if r.status == _RedirectStatusStoping {
				return
			}
			time.Sleep(time.Millisecond)
			continue
		}

		rw.LastTryTime = time.Now()
		res, err := new(http.Client).Do(rw.request)

		rw.LastCostTime = time.Now().Sub(rw.LastTryTime)
		if err != nil {
			rw.LastError = err
			rw.TryCount++
		} else if res.StatusCode != 200 {
			rw.LastError = errors.New(res.Status)
			rw.TryCount++
		} else {
			rw.LastError = nil
			rw.Success = true
		}

		if rw.Success || rw.TryCount >= r.retrytimes {
			r.logRequestWrapper("done", rw)
			rw.clear()
			r.requestPool.Put(rw)
		} else {
			r.putTask(rw, true)
		}
	}

}

func (r *Redirect) logRequestWrapper(cat string, rw *RequestWrapper) {
	log.Printf("[Proxy]\t%s\t%s\n", cat, rw.String())
}

//RequestWrapper request wrapper
type RequestWrapper struct {
	CreatedTime  time.Time
	LastTryTime  time.Time
	LastCostTime time.Duration
	Success      bool
	TryCount     int
	request      *http.Request
	LastError    error
	ID           string
}

func (r *RequestWrapper) clear() {
	r.Success = false
	r.TryCount = 0
	r.request = nil
	r.LastError = nil
	r.LastCostTime = 0
	r.ID = ""
}

func (r *RequestWrapper) String() string {

	requestInfo := ""
	if r.request != nil {
		requestInfo += r.request.Method + " " + r.request.Proto + " " + r.request.URL.String()
	}
	errInfo := ""
	if r.LastError != nil {
		errInfo = r.LastError.Error()
	}
	m := map[string]string{
		"id":           r.ID,
		"createdTime":  r.CreatedTime.String(),
		"lastTryTime":  r.LastTryTime.String(),
		"lastCostTime": fmt.Sprintf("%f", r.LastCostTime.Seconds()),
		"success":      fmt.Sprintf("%v", r.Success),
		"tryCount":     fmt.Sprintf("%v", r.TryCount),
		"request":      requestInfo,
		"lastError":    errInfo,
	}

	j, _ := json.Marshal(m)
	return string(j)
}

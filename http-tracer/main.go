package http_tracer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

type Tracer struct {
	trace *httptrace.ClientTrace
	Debug *TraceDetails
}

func New() Tracer {
	t := Tracer{}
	t.trace, t.Debug = InitTracer()
	return t
}

type TraceDetails struct {
	DNS struct {
		Start    time.Time     `json:"start"`
		End      time.Time     `json:"end"`
		Duration time.Duration `json:"duration"`
		Host     string        `json:"host"`
		Address  []net.IPAddr  `json:"address"`
		Error    error         `json:"error"`
	} `json:"dns"`
	Dial struct {
		Start    time.Time     `json:"start"`
		End      time.Time     `json:"end"`
		Duration time.Duration `json:"duration"`
	} `json:"dial"`
	Connection struct {
		Time     time.Time     `json:"time"`
		Duration time.Duration `json:"duration"`
	} `json:"connection"`
	WroteAllRequestHeaders struct {
		Time     time.Time     `json:"time"`
		Duration time.Duration `json:"duration"`
	} `json:"wrote_all_request_header"`
	WroteAllRequest struct {
		Time     time.Time     `json:"time"`
		Duration time.Duration `json:"duration"`
	} `json:"wrote_all_request"`
	FirstReceivedResponseByte struct {
		Time     time.Time     `json:"time"`
		Duration time.Duration `json:"duration"`
	} `json:"first_received_response_byte"`
}

func InitTracer() (*httptrace.ClientTrace, *TraceDetails) {
	d := &TraceDetails{}

	t := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			t := time.Now()
			log.Println(t, "dns start")
			d.DNS.Start = t
			d.DNS.Host = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t := time.Now()
			log.Println(t, "dns end")
			d.DNS.End = t
			d.DNS.Address = info.Addrs
			d.DNS.Error = info.Err
			d.DNS.Duration = d.DNS.End.Sub(d.DNS.Start) / time.Millisecond
		},
		ConnectStart: func(network, addr string) {
			t := time.Now()
			log.Println(t, "dial start")
			d.Dial.Start = t
		},
		ConnectDone: func(network, addr string, err error) {
			t := time.Now()
			log.Println(t, "dial end")
			d.Dial.End = t
			d.Dial.Duration = d.Dial.End.Sub(d.Dial.Start) / time.Millisecond
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			t := time.Now()
			log.Println(t, "conn time")
			d.Connection.Time = t
			d.Connection.Duration = d.Connection.Time.Sub(d.Dial.End) / time.Millisecond
		},
		WroteHeaders: func() {
			t := time.Now()
			log.Println(t, "wrote all request headers")
			d.WroteAllRequestHeaders.Time = t
			d.WroteAllRequest.Duration = d.WroteAllRequestHeaders.Time.Sub(d.Connection.Time) / time.Millisecond
		},
		WroteRequest: func(wr httptrace.WroteRequestInfo) {
			t := time.Now()
			log.Println(t, "wrote all request")
			d.WroteAllRequest.Time = t
			d.WroteAllRequest.Duration = d.WroteAllRequest.Time.Sub(d.WroteAllRequestHeaders.Time) / time.Millisecond
		},
		GotFirstResponseByte: func() {
			t := time.Now()
			log.Println(t, "first received response byte")
			d.FirstReceivedResponseByte.Time = t
			d.FirstReceivedResponseByte.Duration = d.FirstReceivedResponseByte.Time.Sub(d.WroteAllRequest.Time) / time.Millisecond
		},
	}

	return t, d
}
func (t Tracer) Get(u string) error {
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), t.trace))

	type clientSeSerepetickama struct {
		client        http.Client
		extraResponse *http.Response
	}

	x := clientSeSerepetickama{}
	x.client = http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		x.extraResponse = req.Response
		return errors.New("301: Redirect blocked on purpose")
	}}

	resp, err := x.client.Do(req)
	_, _ = http.DefaultTransport.RoundTrip(req)

	if err != nil {
		fmt.Printf("%v", err)
		//if !strings.Contains(err.Error(), "301: Redirect blocked on purpose") {
		z := err.(*url.Error).Err.Error()
		if z != "301: Redirect blocked on purpose" {
			fmt.Printf("%v", resp)
			return err
		} else {
			fmt.Printf("OK: Reditect blocked on purpose, original response was: %v", x.extraResponse.Status)
		}
	}

	return nil

}

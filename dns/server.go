package dns

import (
	"fmt"

	"github.com/miekg/dns"
)

type Records map[string]string
type onRequestHook func(addr string)

type Server struct {
	addr          string
	zone          string
	records       Records
	onRequestHook onRequestHook

	stopChan chan bool
}

func MakeServer(addr, zone string, records Records) Server {

	return Server{
		addr:    addr,
		zone:    zone,
		records: records,

		onRequestHook: func(addr string) {},

		stopChan: make(chan bool),
	}
}

// Register a function which will be invoked at each request.
func (s *Server) SetOnRequestHook(fn onRequestHook) {
	s.onRequestHook = fn
}

// Start the DNS server
func (s *Server) Start() error {
	dns.HandleFunc(s.zone, s.handleDnsRequest)

	server := &dns.Server{
		Addr: s.addr,
		Net:  "udp",
	}

	err := server.ListenAndServe()

	if err != nil {
		return err
	}

	go func() {
		<-s.stopChan

		server.Shutdown()
		return
	}()

	return nil
}

func (s *Server) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			s.onRequestHook(q.Name)

			ip := s.records[q.Name]
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func (s *Server) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		s.parseQuery(m)
	}

	w.WriteMsg(m)
}

// Stop the DNS server
func (s *Server) Stop() {
	s.stopChan <- true
	close(s.stopChan)
}

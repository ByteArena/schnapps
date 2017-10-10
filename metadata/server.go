package metadata

import (
	"fmt"
	"net/http"

	"github.com/bytearena/schnapps"
	"github.com/bytearena/schnapps/types"
)

type RetrieveVMFn func(id string) *vm.VM

type MetadataHTTPServer struct {
	addr         string
	retrieveVMFn RetrieveVMFn
}

func vmMetadataToString(metadata types.VMMetadata) string {
	var str string

	for k, v := range metadata {
		str += k + "=" + v + ";"
	}

	return str
}

func (server *MetadataHTTPServer) handleMetadataRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, hasId := r.Form["id"]

	if !hasId {
		return
	}

	vm := server.retrieveVMFn(id[0])

	if vm != nil {
		fmt.Fprintf(w, vmMetadataToString(vm.Config.Metadata))
	} else {
		fmt.Fprintf(w, "")
	}
}

func (server *MetadataHTTPServer) Start() error {
	http.HandleFunc("/metadata", server.handleMetadataRequest)

	return http.ListenAndServe(server.addr, nil)
}

// FIXME(sven): implement this
func (server *MetadataHTTPServer) Stop() error {
	return nil
}

func NewServer(addr string, retrieveVMFn RetrieveVMFn) *MetadataHTTPServer {
	return &MetadataHTTPServer{
		addr:         addr,
		retrieveVMFn: retrieveVMFn,
	}
}

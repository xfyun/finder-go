package finderm

import (
	"io"
	"net/http"
	"sync"
)

type httpRpc struct {
	address sync.Map
}

func (r *httpRpc)getServiceAddresses(service ,version string){

}

func (r *httpRpc)HttpCall(service string,apiVersion string,method string,headers map[string]string,body io.Reader)(*http.Response,error){
	http.NewRequest()
}

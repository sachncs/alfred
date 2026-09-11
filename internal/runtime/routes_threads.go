package runtime

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

func (r *LocalRuntime) handleListThreads(w http.ResponseWriter, req *http.Request) {
	threads, nextToken, err := r.ThreadStore().List(50, req.URL.Query().Get("token"))
	if err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, contract.ListThreadsResponse{Threads: threads, NextToken: nextToken}, 200)
}

func (r *LocalRuntime) handleCreateThread(w http.ResponseWriter, req *http.Request) {
	var in contract.CreateThreadRequest
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}
	th := &contract.Thread{
		ID:        contract.ThreadID(fmt.Sprintf("thr-%d", time.Now().UnixNano())),
		Title:     in.Title,
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Metadata:  in.Metadata,
	}
	if err := r.ThreadStore().Create(th); err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	RouteJSON(w, contract.CreateThreadResponse{Thread: *th}, 201)
}

func (r *LocalRuntime) handleGetThread(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	th, err := r.ThreadStore().Get(id)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}
	RouteJSON(w, th, 200)
}

func (r *LocalRuntime) handleDeleteThread(w http.ResponseWriter, req *http.Request) {
	id := contract.ThreadID(req.PathValue("id"))
	if err := r.ThreadStore().Delete(id); err != nil {
		RouteError(w, contract.CodeInternal, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

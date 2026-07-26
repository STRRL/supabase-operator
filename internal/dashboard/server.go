package dashboard

import (
	"encoding/json"
	"net/http"
	"sort"

	"sigs.k8s.io/controller-runtime/pkg/client"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
)

type projectListResponse struct {
	Projects []projectResponse `json:"projects"`
}

type projectResponse struct {
	Namespace       string              `json:"namespace"`
	Name            string              `json:"name"`
	ProjectID       string              `json:"projectId"`
	Phase           string              `json:"phase"`
	Message         string              `json:"message"`
	ReadyComponents int                 `json:"readyComponents"`
	TotalComponents int                 `json:"totalComponents"`
	Components      []componentResponse `json:"components"`
	UpdatedAt       string              `json:"updatedAt,omitempty"`
}

type componentResponse struct {
	Name          string `json:"name"`
	Phase         string `json:"phase"`
	Ready         bool   `json:"ready"`
	ReadyReplicas int32  `json:"readyReplicas"`
	Replicas      int32  `json:"replicas"`
}

type handler struct {
	reader client.Reader
}

func NewHandler(reader client.Reader) http.Handler {
	dashboardHandler := &handler{reader: reader}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/projects", dashboardHandler.listProjects)
	mux.HandleFunc("GET /", dashboardHandler.index)
	return mux
}

func (h *handler) listProjects(responseWriter http.ResponseWriter, request *http.Request) {
	projects := &supabasev1alpha1.SupabaseProjectList{}
	if err := h.reader.List(request.Context(), projects); err != nil {
		http.Error(responseWriter, "failed to list projects", http.StatusInternalServerError)
		return
	}

	response := projectListResponse{
		Projects: make([]projectResponse, 0, len(projects.Items)),
	}
	for index := range projects.Items {
		response.Projects = append(response.Projects, newProjectResponse(&projects.Items[index]))
	}
	sort.Slice(response.Projects, func(left, right int) bool {
		leftName := response.Projects[left].Namespace + "/" + response.Projects[left].Name
		rightName := response.Projects[right].Namespace + "/" + response.Projects[right].Name
		return leftName < rightName
	})

	responseWriter.Header().Set("Cache-Control", "no-store")
	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(response); err != nil {
		return
	}
}

func (h *handler) index(responseWriter http.ResponseWriter, _ *http.Request) {
	responseWriter.Header().Set("Cache-Control", "no-cache")
	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = responseWriter.Write(indexHTML)
}

func newProjectResponse(project *supabasev1alpha1.SupabaseProject) projectResponse {
	components := []componentResponse{
		newComponentResponse("Kong", project.Status.Components.Kong),
		newComponentResponse("Auth", project.Status.Components.Auth),
		newComponentResponse("Realtime", project.Status.Components.Realtime),
		newComponentResponse("PostgREST", project.Status.Components.PostgREST),
		newComponentResponse("Storage", project.Status.Components.StorageAPI),
		newComponentResponse("Meta", project.Status.Components.Meta),
		newComponentResponse("Studio", project.Status.Components.Studio),
	}

	readyComponents := 0
	for _, component := range components {
		if component.Ready {
			readyComponents++
		}
	}

	response := projectResponse{
		Namespace:       project.Namespace,
		Name:            project.Name,
		ProjectID:       project.Spec.ProjectID,
		Phase:           project.Status.Phase,
		Message:         project.Status.Message,
		ReadyComponents: readyComponents,
		TotalComponents: len(components),
		Components:      components,
	}
	if project.Status.LastReconcileTime != nil {
		response.UpdatedAt = project.Status.LastReconcileTime.UTC().Format("2006-01-02T15:04:05Z")
	}
	return response
}

func newComponentResponse(name string, component supabasev1alpha1.ComponentStatus) componentResponse {
	return componentResponse{
		Name:          name,
		Phase:         component.Phase,
		Ready:         component.Ready,
		ReadyReplicas: component.ReadyReplicas,
		Replicas:      component.Replicas,
	}
}

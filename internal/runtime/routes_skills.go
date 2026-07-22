package runtime

import (
	"net/http"
)

func (r *LocalRuntime) handleListSkills(w http.ResponseWriter, req *http.Request) {
	skills, err := r.SkillsLoader().ListSkills()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if skills == nil {
		skills = []SkillInfo{}
	}
	RouteJSON(w, map[string]any{"skills": skills}, 200)
}

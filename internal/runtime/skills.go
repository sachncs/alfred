package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// SkillInfo represents a loaded skill definition.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Content     string `json:"content"`
}

// SkillsLoader discovers and loads skills from the workspace.
type SkillsLoader struct {
	workspaceRoot string
}

// NewSkillsLoader creates a loader rooted at the current directory.
// ponytail: zero-arg constructor so HTTPRuntime can hold one unconditionally;
// call SetWorkspaceRoot before use to point at the actual workspace.
func NewSkillsLoader() *SkillsLoader {
	return &SkillsLoader{workspaceRoot: "."}
}

// SetWorkspaceRoot sets the directory skills are loaded from.
func (l *SkillsLoader) SetWorkspaceRoot(root string) {
	l.workspaceRoot = root
}

// ListSkills returns all available skills in the workspace.
func (l *SkillsLoader) ListSkills() ([]SkillInfo, error) {
	skillsDir := filepath.Join(l.workspaceRoot, ".alfred", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SkillInfo{}, nil
		}
		return nil, err
	}

	var skills []SkillInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		skills = append(skills, SkillInfo{
			Name: name,
			Path: filepath.Join(skillsDir, e.Name()),
		})
	}
	return skills, nil
}

// LoadSkill returns the full content of a skill by name.
func (l *SkillsLoader) LoadSkill(name string) (*SkillInfo, error) {
	path := filepath.Join(l.workspaceRoot, ".alfred", "skills", name+".md")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &SkillInfo{
		Name:    name,
		Path:    path,
		Content: string(content),
	}, nil
}

// ListSkillsJSON returns skills as JSON.
func (l *SkillsLoader) ListSkillsJSON() (json.RawMessage, error) {
	skills, err := l.ListSkills()
	if err != nil {
		return nil, err
	}
	return json.Marshal(skills)
}

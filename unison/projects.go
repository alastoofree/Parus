// pars/unison/projects.go

package unison

import (
	"fmt"
	"os"
	"path/filepath"
	"strings")

// Project — тема (папка на диске)
type Project struct {
	ID   string
	Meta MetaFrame}

// CreateProject создаёт новую тему на диске
func CreateProject(title string, fileOrderLen int) (Project, error) {
id := NewULID()
dir := filepath.Join(DataDir, id)
if err := os.MkdirAll(dir, 0755); err != nil {
	return Project{}, fmt.Errorf("mkdir %s: %w", dir, err)}
meta := MetaFrame{
	SlotOrderLen: fileOrderLen,
	FileOrderLen: fileOrderLen,
	Frame: Frame{
		Index: ToGauss(0, fileOrderLen),
		Pairs: []SlotPair{
			{TypeIndex: "n", Left: "id",         Right: id},
			{TypeIndex: "m", Left: "title",      Right: title},
			{TypeIndex: "l", Left: "a-sort",     Right: ""},
			{TypeIndex: "r", Left: "o-lens",     Right: ""},
			{TypeIndex: "h", Left: "ranks",      Right: ""},
			{TypeIndex: "g", Left: "tabs",       Right: ""},
			{TypeIndex: "k", Left: "west",       Right: "tags"},
			{TypeIndex: "c", Left: "nord",       Right: "links"},
			{TypeIndex: "s", Left: "deep-nord",  Right: ""},
			{TypeIndex: "z", Left: "deep-south", Right: ""},
			{TypeIndex: "t", Left: "graph",      Right: ""},
			{TypeIndex: "d", Left: "list",       Right: ""},
			{TypeIndex: "b", Left: "chrono",     Right: ""},
			{TypeIndex: "p", Left: "reserve",    Right: ""},
			{TypeIndex: "f", Left: "reserve",    Right: ""},
			{TypeIndex: "v", Left: "reserve",    Right: ""},
			{TypeIndex: "w", Left: "reserve",    Right: ""},
			{TypeIndex: "u", Left: "reserve",    Right: ""},
			{TypeIndex: "o", Left: "reserve",    Right: ""},
			{TypeIndex: "a", Left: "reserve",    Right: ""},
			{TypeIndex: "e", Left: "counts",     Right: "0 0"},
			{TypeIndex: "i", Left: "all-titles", Right: ""},
			{TypeIndex: "y", Left: "tags",       Right: ""},
			{TypeIndex: "j", Left: "links",      Right: ""},
			{TypeIndex: "q", Left: "includes",   Right: ""},
			{TypeIndex: "x", Left: "product",    Right: ""},},},}
if err := WriteMetaFrame(id, meta); err != nil {
	return Project{}, fmt.Errorf("write metaframe: %w", err)}
return Project{ID: id, Meta: meta}, nil }
// LoadProject загружает тему из папки
func LoadProject(id string) (Project, error) {
metaMu.Lock()
defer metaMu.Unlock() 
dir := filepath.Join(DataDir, id)
entries, err := os.ReadDir(dir)
if err != nil {return Project{}, fmt.Errorf("readdir %s: %w", dir, err)}
for _, e := range entries {
	name := e.Name()
	if !strings.HasSuffix(name, ".uic") {continue}
	stem := strings.TrimSuffix(name, ".uic")
	if stem == "" || strings.TrimLeft(stem, "0") != "" {continue}
	meta, err := ParseMetaFrame(filepath.Join(dir, name))
	if err != nil {return Project{}, err}
	return Project{ID: id, Meta: meta}, nil}
return Project{}, fmt.Errorf("metaframe not found in %s", dir) }
// ProjectConfig — параметры конфигурации темы
type ProjectConfig struct {
	Title     string
	ASort     string
	OLens     string
	Ranks     string
	Tabs      string
	West      string
	Nord      string
	Graph     string
	List      string
	Chrono    string
	DeepNord  string
	DeepSouth string}
// UpdateProject обновляет конфигурационные параметры темы
func UpdateProject(projectID string, p Project, cfg ProjectConfig) (Project, error) {
updates := map[string]string{
	"a-sort":     cfg.ASort,
	"o-lens":     cfg.OLens,
	"ranks":      cfg.Ranks,
	"tabs":       cfg.Tabs,
	"west":       cfg.West,
	"nord":       cfg.Nord,
	"graph":      cfg.Graph,
	"list":       cfg.List,
	"chrono":     cfg.Chrono,
	"deep-nord":  cfg.DeepNord,
	"deep-south": cfg.DeepSouth,}
for i, pair := range p.Meta.Frame.Pairs {
	if pair.Left == "title" {p.Meta.Frame.Pairs[i].Right = cfg.Title; continue}
	if v, ok := updates[pair.Left]; ok {p.Meta.Frame.Pairs[i].Right = v}}
if err := WriteMetaFrame(projectID, p.Meta); err != nil {
	return Project{}, fmt.Errorf("write metaframe: %w", err)}
return p, nil }
// ListProjects возвращает все темы из DataDir
func ListProjects() ([]Project, error) {
entries, err := os.ReadDir(DataDir)
if err != nil {
	if os.IsNotExist(err) {return []Project{}, nil}
	return nil, fmt.Errorf("readdir %s: %w", DataDir, err)}
var projects []Project
for _, e := range entries {
	if !e.IsDir() {continue}
	p, err := LoadProject(e.Name())
	if err != nil {continue}
	projects = append(projects, p)}
return projects, nil }
// parus/unison/frames.go

package unison

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sort"
	"time")

// DataDir — корневая папка всех проектов
const DataDir = "./data"
// Frame — упорядоченный список пар слотов (продукций)
type Frame struct {
	Index string
	Pairs []SlotPair}
// MetaFrame — конфигурация темы
type MetaFrame struct {
	Frame
	SlotOrderLen int
	FileOrderLen int}
// slotKey - карта фрейма
var slotKey = map[string]string{
	"n": "id",           "m": "modified",
    "l": "title",        "r": "alias",
    "h": "tags",         "g": "content",
	"k": "content-type", "c": "links",
    "s": "includes",     "z": "icon",
    "t": "color",        "d": "rank",
	"b": "tab",          "p": "i-sort",
    "f": "west",         "v": "nord",
    "w": "graph",        "u": "list",
	"o": "match-nord",   "a": "match-south",
	"e": "chrono",       "i": "i-chrono",
	"y": "dict",         "j": "i-dict",
    "q": "show-product", "x": "product",}
// Имена слотов фрейма
const (
	FrameSlotLinks    = "links"
	FrameSlotTags     = "tags"
	FrameSlotAlias    = "alias"
	FrameSlotIncludes = "includes"
	FrameSlotProduct  = "product")
var slotLetter = map[string]string{}

func init() { for k, v := range slotKey { slotLetter[v] = k } }
// FramePairRight возвращает значение правого слота по имени левого
func FramePairRight(f Frame, key string) string {
for _, p := range f.Pairs {
	if p.Left == key {
		return p.Right} }
return "" }
// WriteFrame записывает фрейм в .uic файл
func WriteFrame(projectID string, frame Frame) error {
dir := filepath.Join(DataDir, projectID)
path := filepath.Join(dir, frame.Index+".uic")
var sb strings.Builder
for _, pair := range frame.Pairs {
	letter := slotLetter[pair.Left]
	if letter == "" { letter = pair.TypeIndex }
	fmt.Fprintf(&sb, "l-%s %s\n", letter, pair.Left)
	fmt.Fprintf(&sb, "r-%s %s\n", letter, strings.ReplaceAll(pair.Right, "\n", `\n`))}
return os.WriteFile(path, []byte(sb.String()), 0644) }
// WriteMetaFrame записывает метафрейм в .uic файл
func WriteMetaFrame(projectID string, meta MetaFrame) error {
metaMu.Lock()
defer metaMu.Unlock()
dir := filepath.Join(DataDir, projectID)
filename := strings.Repeat("0", meta.FileOrderLen) + ".uic"
path := filepath.Join(dir, filename)
var sb strings.Builder
for _, pair := range meta.Frame.Pairs {
	fmt.Fprintf(&sb, "l-%s %s\n", pair.TypeIndex, pair.Left)
	fmt.Fprintf(&sb, "r-%s %s\n", pair.TypeIndex, pair.Right)}
return os.WriteFile(path, []byte(sb.String()), 0644) }
// UniqueTitle возвращает уникальный титул в рамках темы
// skipIndex — индекс фрейма который пропускаем при проверке (при редактировании)
// При создании передавать skipIndex = ""
func UniqueTitle(projectID, title, skipTitle string, meta MetaFrame) string {
allTitles := FramePairRight(meta.Frame, "all-titles")
titles := ParseTokenList(allTitles)
base    := title
counter := 2
for {
    taken := false
    for _, t := range titles {
        if t == title && t != skipTitle {taken = true; break}}
    if !taken {break}
    title = fmt.Sprintf("%s-%d", base, counter)
    counter++}
return title }
// LoadFrame загружает один фрейм темы
func LoadFrame(projectID, frameIndex string) (Frame, error) {
path := filepath.Join(DataDir, projectID, frameIndex+".uic")
f, err := ParseFrame(path)
if err != nil {return Frame{}, err}
f.Index = frameIndex
return f, nil }
// CreateFrame создаёт новый фрейм в теме
func CreateFrame(projectID string, meta MetaFrame, title string) (Frame, error) {
dir := filepath.Join(DataDir, projectID)
entries, err := os.ReadDir(dir)
if err != nil {
	return Frame{}, fmt.Errorf("readdir %s: %w", dir, err)}
frameCount := 0
for _, e := range entries {
	name := e.Name()
	if !strings.HasSuffix(name, ".uic") {continue}
	stem := strings.TrimSuffix(name, ".uic")
	if stem != "" && strings.TrimLeft(stem, "0") != "" {frameCount++}}
maxFrames := 1
for i := 0; i < meta.FileOrderLen; i++ {maxFrames *= 18}
if frameCount >= maxFrames {
	return Frame{}, fmt.Errorf("frame limit reached: max %d frames for order %d", maxFrames, meta.FileOrderLen)}
title = UniqueTitle(projectID, title, "", meta)
frameIndex := ToGauss(frameCount, meta.FileOrderLen)
now := time.Now().Format("2006-01-02")
frame := Frame{
	Index: frameIndex,
	Pairs: []SlotPair{
		{TypeIndex: "n", Left: "id",           Right: NewULID()},
		{TypeIndex: "m", Left: "modified",     Right: now},
		{TypeIndex: "l", Left: "title",        Right: title},
		{TypeIndex: "r", Left: "alias",        Right: ""},
		{TypeIndex: "h", Left: "tags",         Right: ""},
		{TypeIndex: "g", Left: "content",      Right: ""},
		{TypeIndex: "k", Left: "content-type", Right: ""},
		{TypeIndex: "c", Left: "links",        Right: ""},
		{TypeIndex: "s", Left: "includes",     Right: ""},
		{TypeIndex: "z", Left: "icon",         Right: ""},
		{TypeIndex: "t", Left: "color",        Right: ""},
		{TypeIndex: "d", Left: "rank",         Right: ""},
		{TypeIndex: "b", Left: "tab",          Right: ""},
		{TypeIndex: "p", Left: "i-sort",       Right: ""},
        {TypeIndex: "f", Left: "west",         Right: ""},
		{TypeIndex: "v", Left: "nord",         Right: ""},
		{TypeIndex: "w", Left: "graph",        Right: ""},
		{TypeIndex: "u", Left: "list",         Right: ""},
		{TypeIndex: "o", Left: "match-nord",   Right: ""},
		{TypeIndex: "a", Left: "match-south",  Right: ""},
		{TypeIndex: "e", Left: "chrono",       Right: ""},
		{TypeIndex: "i", Left: "i-chrono",     Right: ""},
		{TypeIndex: "y", Left: "dict",         Right: ""},
		{TypeIndex: "j", Left: "i-dict",       Right: ""},
		{TypeIndex: "q", Left: "show-product", Right: ""},
		{TypeIndex: "x", Left: "product",      Right: ""},},}
if err := WriteFrame(projectID, frame); err != nil {
	return Frame{}, fmt.Errorf("write frame: %w", err)}
_ = UpdateFrameTitle(projectID, frame.Index, title)
_ = Count(projectID, 1, 0)
return frame, nil }
// UpdateFrameTitle обновляет титул фрейма в all-titles мета-фрейма
func UpdateFrameTitle(projectID, frameIndex, newTitle string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
pos := GaussOrder(frameIndex)
allTitles := FramePairRight(p.Meta.Frame, "all-titles")
tokens := ParseTokenList(allTitles)
for len(tokens) <= pos {tokens = append(tokens, "")}
tokens[pos] = newTitle
var sb strings.Builder
for i, t := range tokens {
	if i > 0 {sb.WriteString(" ")}
	sb.WriteString(quoteIfSpace(t))}
for i, pair := range p.Meta.Frame.Pairs {
	if pair.Left == "all-titles" {
		p.Meta.Frame.Pairs[i].Right = sb.String(); break}}
return WriteMetaFrame(projectID, p.Meta) }
// ListFrames
func ListFrames(projectID string, meta MetaFrame) ([]Frame, error) {
dir := filepath.Join(DataDir, projectID)
entries, err := os.ReadDir(dir)
if err != nil {return nil, fmt.Errorf("readdir %s: %w", dir, err)}
var frames []Frame
for _, e := range entries {
	name := e.Name()
	stem := strings.TrimSuffix(name, ".uic")
    if !strings.HasSuffix(name, ".uic") || strings.TrimLeft(stem, "0") == "" {continue}
	frame, err := ParseFrame(filepath.Join(dir, name))
    if err != nil {continue}
    frame.Index = strings.TrimSuffix(name, ".uic")
    frames = append(frames, frame)}
return frames, nil }
// ListTags возвращает все фреймы-теги темы
func ListTags(projectID string, meta MetaFrame) ([]Frame, error) {
frames, err := ListFrames(projectID, meta)
if err != nil {return nil, err}
var tags []Frame
for _, f := range frames {
	if FramePairRight(f, "rank") != "" {tags = append(tags, f)}}
return tags, nil }
// TitleMap
func TitleMap(meta MetaFrame) map[string]string {
allTitles := FramePairRight(meta.Frame, "all-titles")
tokens := ParseTokenList(allTitles)
result := make(map[string]string, len(tokens))
for i, title := range tokens {
	idx := ToGauss(i, meta.FileOrderLen)
	result[idx] = title}
return result }
// updateMetaSlots записывает map слот→значение в мета-фрейм
func updateMetaSlots(meta *MetaFrame, updates map[string]string) {
for i, pair := range meta.Frame.Pairs {
	if v, ok := updates[pair.Left]; ok {
		meta.Frame.Pairs[i].Right = v}} }
// ListMetaFrames читает все фреймы мета-базы
func ListMetaFrames() ([]GlobalFrame, error) {
p, err := LoadProject(MetaBaseName)
if err != nil {return nil, err}
frames, err := ListFrames(MetaBaseName, p.Meta)
if err != nil {return nil, err}
metaIdx := strings.Repeat("0", MetaBaseFileLen)
var result []GlobalFrame
for _, f := range frames {
    if f.Index == metaIdx {continue}
    alias := []string{}
    if raw := FramePairRight(f, "alias"); raw != "" {alias = ParseTokenList(raw)}
    result = append(result, GlobalFrame{
        Index:   f.Index,
        Title:   FramePairRight(f, "title"),
        Rank:    FramePairRight(f, "rank"),
        Aliases: alias,})}
return result, nil}
// CreateMetaFrames создаёт новый фрейм-тег в мета-базе
func CreateMetaFrames(title, rank string) (GlobalFrame, error) {
existing, _ := ListMetaFrames()
for _, gf := range existing {
	if strings.EqualFold(gf.Title, title) {return gf, nil}}
p, err := LoadProject(MetaBaseName)
if err != nil {return GlobalFrame{}, err}
frame, err := CreateFrame(MetaBaseName, p.Meta, title)
if err != nil {return GlobalFrame{}, err}
now := time.Now().Format("2006-01-02")
for i, pair := range frame.Pairs {
	switch pair.Left {
	case "rank":     frame.Pairs[i].Right = rank
	case "modified": frame.Pairs[i].Right = now}}
if err := WriteFrame(MetaBaseName, frame); err != nil {return GlobalFrame{}, err}
if rank != "" {_ = Count(MetaBaseName, 0, 1)}
return GlobalFrame{
	Index: frame.Index,
	Title: title,
	Rank:  rank,}, nil }
// CreateSatellite создаёт фрейм-спутник с типом json для продукций
func CreateSatellite(projectID string, meta MetaFrame) (Frame, error) {
	sat, err := CreateFrame(projectID, meta, "")
	if err != nil {return Frame{}, err}
	for i, p := range sat.Pairs {
		if p.Left == "content-type" {sat.Pairs[i].Right = "json"}
		if p.Left == "content"      {sat.Pairs[i].Right = "{}"}}
	if err := WriteFrame(projectID, sat); err != nil {return Frame{}, err}
	return sat, nil }
// GetSatellite загружает фрейм-спутник по индексу из слота x
func GetSatellite(projectID string, frame Frame) (Frame, error) {
	idx := FramePairRight(frame, "product")
	if idx == "" {return Frame{}, fmt.Errorf("no satellite")}
	return LoadFrame(projectID, idx) }
// GetChildren возвращает дочерние фреймы тега в правильном порядке
func GetChildren(projectID, frameIndex string, meta MetaFrame, iSort, aSort string) ([]map[string]string, error) {
children := parseMapSlot(meta.Frame, "tags", meta.FileOrderLen)[frameIndex]
if len(children) == 0 {return []map[string]string{}, nil}
titleMap := TitleMap(meta)
if iSort != "" {
	order := map[string]int{}
	for i, idx := range ParseTokenList(iSort) {order[idx] = i}
	sort.Slice(children, func(i, j int) bool {
		oi := order[children[i]]; if _, ok := order[children[i]]; !ok {oi = 9999}
		oj := order[children[j]]; if _, ok := order[children[j]]; !ok {oj = 9999}
		return oi < oj})}  else if aSort != "" {
	charOrder := BuildCharOrder(aSort)
	sort.Slice(children, func(i, j int) bool {
		return CompareByASort(titleMap[children[i]], titleMap[children[j]], charOrder) < 0})}
result := make([]map[string]string, 0, len(children))
for _, idx := range children {
	result = append(result, map[string]string{"index": idx, "title": titleMap[idx]})}
return result, nil }
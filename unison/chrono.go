// parus/unison/chrono.go

package unison

import (
	"encoding/json"
	"fmt"
	"strings")

// ChronoRow — строка хронологической таблицы
type ChronoRow struct {
	FrameIndex string            `json:"frameIndex"`
	FrameTitle string            `json:"frameTitle"`
	DateVal    string            `json:"dateVal"`
	DateKey    string            `json:"dateKey"`
	Extras     map[string]string `json:"extras"`}

// GetChrono возвращает хронологическую таблицу для фрейма
func GetChrono(projectID, frameIndex string) ([]ChronoRow, error) {
frame, err := LoadFrame(projectID, frameIndex)
if err != nil {return nil, fmt.Errorf("load frame: %w", err)}
chronoVal  := FramePairRight(frame, "chrono")
iChronoVal := FramePairRight(frame, "i-chrono")
if chronoVal == "" || iChronoVal == "" {return []ChronoRow{}, nil}
retro  := strings.HasPrefix(chronoVal, "-")
tokens := ParseTokenList(iChronoVal)
if len(tokens) == 0 {return []ChronoRow{}, nil}
extraCols := tokens[1:]
p, err := LoadProject(projectID)
if err != nil {return nil, fmt.Errorf("load project: %w", err)}
titleMap   := TitleMap(p.Meta)
chronoList := ParseTokenList(FramePairRight(p.Meta.Frame, "chrono"))
var rows []ChronoRow
for _, fid := range chronoList {
	f, err2 := LoadFrame(projectID, fid)
	if err2 != nil {continue}
	iChrono := ParseTokenList(FramePairRight(f, "i-chrono"))
	if len(iChrono) == 0 {continue}
	sat, err3 := LoadFrame(projectID, iChrono[0])
	if err3 != nil {continue}
	contentStr := FramePairRight(sat, "content")
	var data map[string]string
	if json.Unmarshal([]byte(contentStr), &data) != nil {continue}
	dec := json.NewDecoder(strings.NewReader(contentStr))
	dec.Token()
	dateKey := ""
	if tok, err4 := dec.Token(); err4 == nil {
		if s, ok := tok.(string); ok {dateKey = s}}
	if dateKey == "" {continue}
	extras := make(map[string]string)
	for _, col := range extraCols {
		if v, ok := data[col]; ok {extras[col] = v}}
	rows = append(rows, ChronoRow{
		FrameIndex: fid,
		FrameTitle: titleMap[fid],
		DateVal:    data[dateKey],
		DateKey:    dateKey,
		Extras:     extras,})}
if retro {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {rows[i], rows[j] = rows[j], rows[i]}
} else {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].DateVal < rows[j-1].DateVal; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]}}}
return rows, nil }
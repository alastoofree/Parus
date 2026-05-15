// pars/unison/union.go

package unison

import (
	"fmt"
	"regexp"
	"sort"
	"strings")

// UnionFrames объединяет группу фреймов с одинаковым базовым титулом
func UnionFrames(projectID, anyIndex string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
f, err := LoadFrame(projectID, anyIndex)
if err != nil {return fmt.Errorf("load frame: %w", err)}
title    := FramePairRight(f, "title")
reNum := regexp.MustCompile(`-\d+$`)
baseTitle := reNum.ReplaceAllString(title, "")
allFrames, err := ListFrames(projectID, p.Meta)
if err != nil {return fmt.Errorf("list frames: %w", err)}
var group []Frame
for _, fr := range allFrames {
	t := FramePairRight(fr, "title")
	if t == baseTitle || reNum.ReplaceAllString(t, "") == baseTitle {
		group = append(group, fr)}}
if len(group) < 2 {return fmt.Errorf("менее двух фреймов в группе")}
sort.Slice(group, func(i, j int) bool {
	return GaussOrder(group[i].Index) < GaussOrder(group[j].Index)})
// Главный — с базовым титулом, иначе первый
var main Frame
var rest []Frame
for _, fr := range group {
	if FramePairRight(fr, "title") == baseTitle {main = fr} else {rest = append(rest, fr)}}
if main.Index == "" {main = group[0]; rest = group[1:]}
// Слоты для объединения списков
listSlots := map[string]bool{"tags": true, "links": true, "includes": true, "alias": true, "product": true}
// Слоты берём из первого непустого
singleSlots := []string{"icon", "color", "tab", "i-sort", "chrono", "i-chrono", "dict", "i-dict", "show-product", "west", "nord", "graph", "list", "match-nord", "match-south"}
// Собираем данные из всей группы
contentParts := []string{}
listData     := make(map[string][]string)
rankMax      := 0
extraSlots   := make(map[string]string) // дополнительные слоты
for _, fr := range group {
	if c := FramePairRight(fr, "content"); c != "" {
		contentParts = append(contentParts, c)}
	for _, pair := range fr.Pairs {
		if listSlots[pair.Left] {
			for _, ref := range ParseTokenList(pair.Right) {
				listData[pair.Left] = AppendUnique(listData[pair.Left], ref)}}
		if pair.Left == "rank" {
			if n := parseIntDefault(pair.Right, 0); n > rankMax {rankMax = n}}
		// Дополнительные слоты
        if !frameSystemSlots[pair.Left] {
            if extraSlots[pair.Left] == "" {
                extraSlots[pair.Left] = pair.Right}}}}
// Обновляем главный фрейм
singleVals := make(map[string]string)
for _, slot := range singleSlots {
	for _, fr := range group {
		if v := FramePairRight(fr, slot); v != "" {singleVals[slot] = v; break}}}
for i, pair := range main.Pairs {
	switch {
	case pair.Left == "content": main.Pairs[i].Right = strings.Join(contentParts, "<br>")
	case pair.Left == "rank": if rankMax > 0 {main.Pairs[i].Right = fmt.Sprintf("%d", rankMax)}
	case listSlots[pair.Left]: main.Pairs[i].Right = strings.Join(listData[pair.Left], " ")
default: if v, ok := singleVals[pair.Left]; ok {main.Pairs[i].Right = v}}}
// Добавляем дополнительные слоты которых нет в главном
for left, right := range extraSlots {
found := false
for _, pair := range main.Pairs {if pair.Left == left {found = true; break}}
if !found {
	userCount := 0
	for _, pair := range main.Pairs {
		if len(pair.TypeIndex) == p.Meta.SlotOrderLen {userCount++}}
	main.Pairs = append(main.Pairs, SlotPair{
		TypeIndex: ToGauss(userCount, p.Meta.SlotOrderLen),
		Left:      left,
		Right:     right})}}
if err := WriteFrame(projectID, main); err != nil {return err}
// Обнуляем остальные
xCount    := 0
xIndexMap := make(map[string]string)
for _, fr := range rest {
	xIndexMap[fr.Index] = main.Index
	xTitle := "X-FRAME"
    if xCount > 0 {xTitle = fmt.Sprintf("X-FRAME-%d", xCount)}
    xTitle = UniqueTitle(projectID, xTitle, "", p.Meta)
	xCount++
	for i, pair := range fr.Pairs {
	    switch pair.Left {
	    case "title": fr.Pairs[i].Right = xTitle
	    case "content","tags","links","includes","alias","product","rank","icon","color",
            "tab","chrono","west","nord","graph","list","match-nord","match-south":
		    fr.Pairs[i].Right = ""
	    default:
		if len(pair.TypeIndex) == p.Meta.SlotOrderLen {fr.Pairs[i].Right = ""}}}
	_ = WriteFrame(projectID, fr)
	_ = UpdateFrameTitle(projectID, fr.Index, xTitle)}
// Обновляем ссылки во всех фреймах
for _, fr := range allFrames {
	changed := false
	for i, pair := range fr.Pairs {
		switch pair.Left {
		case "tags","links","includes","alias","product":
			refs := ParseTokenList(pair.Right)
			var newRefs []string
			for _, ref := range refs {
				if newIdx, ok := xIndexMap[ref]; ok {newRefs = append(newRefs, newIdx)
				} else {newRefs = append(newRefs, ref)}}
			newVal := strings.Join(newRefs, " ")
			if newVal != pair.Right {fr.Pairs[i].Right = newVal; changed = true}
		case "content":
			c := pair.Right
			for oldIdx := range xIndexMap {
				c = strings.ReplaceAll(c, "{{"+oldIdx+"}}", "{{"+main.Index+"}}")
				c = strings.ReplaceAll(c, "[["+oldIdx+"]]", "[["+main.Index+"]]")}
			if c != pair.Right {fr.Pairs[i].Right = c; changed = true}}}
	if changed {_ = WriteFrame(projectID, fr)}}
return RebuildTitles(projectID) }
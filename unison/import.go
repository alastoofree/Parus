// pars/unison/import.go

package unison

import (
	"fmt"
	"os"
	"path/filepath"
    "encoding/json"
	"strings")

// CloneFrame клонирует фрейм
// Копирует все слоты кроме id и modified — они генерируются заново
// func CloneFrame(srcProjectID, srcIndex, dstProjectID string, dstMeta MetaFrame) (Frame, error) {
func CloneFrame(srcProjectID, srcIndex, dstProjectID string, dstMeta MetaFrame, indexMap map[string]string) (Frame, error) {
src, err := LoadFrame(srcProjectID, srcIndex)
if err != nil {return Frame{}, fmt.Errorf("load frame: %w", err)}
title := FramePairRight(src, "title")
dst, err := CreateFrame(dstProjectID, dstMeta, title)
if err != nil {return Frame{}, fmt.Errorf("create frame: %w", err)}
// Копируем слоты кроме системных
for _, sp := range src.Pairs {
	switch sp.Left {
	case "id", "modified", "title":
		continue}
	for i, dp := range dst.Pairs {
		if dp.Left == sp.Left {
			dst.Pairs[i].Right = sp.Right; break}}}
// Добавляем нестандартные слоты
for _, sp := range src.Pairs {
	if frameSystemSlots[sp.Left] {continue}
	found := false
	for _, dp := range dst.Pairs {
		if dp.Left == sp.Left {found = true; break}}
	if !found {
		userCount := 0
		for _, p := range dst.Pairs {
			if len(p.TypeIndex) == dstMeta.FileOrderLen {userCount++}}
		dst.Pairs = append(dst.Pairs, SlotPair{
			TypeIndex: ToGauss(userCount, dstMeta.FileOrderLen),
			Left:      sp.Left,
			Right:     sp.Right})}}
// При клонировании между разными базами — иконки
if srcProjectID != dstProjectID {
    // Клонируем фрейм-иконку если иконка задана индексом (без точки)
	icon := FramePairRight(dst, "icon")
	if icon != "" && !strings.Contains(icon, ".") {
		iconFrame, err2 := LoadFrame(srcProjectID, icon)
		if err2 == nil {
			newIcon, err3 := CreateFrame(dstProjectID, dstMeta, FramePairRight(iconFrame, "title"))
			if err3 == nil {
				for i, p := range newIcon.Pairs {
					for _, sp := range iconFrame.Pairs {
						if sp.Left == p.Left {newIcon.Pairs[i].Right = sp.Right; break}}}
				_ = WriteFrame(dstProjectID, newIcon)
				for i, p := range dst.Pairs {
					if p.Left == "icon" {dst.Pairs[i].Right = newIcon.Index; break}}}}}
	// Копируем файл иконки если есть
    if icon != "" && strings.Contains(icon, ".") {
        srcFile := filepath.Join(DataDir, srcProjectID, "img", icon)
        dstFile := filepath.Join(DataDir, dstProjectID, "img", icon)
        if data, err2 := os.ReadFile(srcFile); err2 == nil {
            dstDir := filepath.Join(DataDir, dstProjectID, "img")
            if err3 := os.MkdirAll(dstDir, 0755); err3 == nil {
                _ = os.WriteFile(dstFile, data, 0644)}}}}
if err := WriteFrame(dstProjectID, dst); err != nil {
	return Frame{}, fmt.Errorf("write frame: %w", err)}
return dst, nil }
// 
func ImportFrames(srcProjectID, dstProjectID string, indices []string) (int, error) {
dstProject, err := LoadProject(dstProjectID)
if err != nil {return 0, fmt.Errorf("load dst project: %w", err)}
frames, err := CollectImportFrames(srcProjectID, indices)
if err != nil {return 0, err}
dstTitleMap := TitleMap(dstProject.Meta)
dstTitles   := make(map[string]string)
for idx, t := range dstTitleMap {
	dstTitles[strings.ToLower(t)] = idx}
indexMap := make(map[string]string)
cloned   := []Frame{}
for _, f := range frames {
	dstProject, err = LoadProject(dstProjectID)
	if err != nil {continue}
	srcTitle := strings.ToLower(FramePairRight(f, "title"))
	if existIdx, ok := dstTitles[srcTitle]; ok {
		indexMap[f.Index] = existIdx
		continue}
	dst, err := CloneFrame(srcProjectID, f.Index, dstProjectID, dstProject.Meta, map[string]string{})
	if err != nil {continue}
	indexMap[f.Index] = dst.Index
	cloned = append(cloned, dst)}
for _, f := range cloned {
	changed := false
	for i, pair := range f.Pairs {
		switch pair.Left {
        case "alias", "tags", "links", "includes", "product", "show-product":
            refs := ParseTokenList(pair.Right)
            var newRefs []string
            for _, ref := range refs {
                if newIdx, ok := indexMap[ref]; ok {
                    if pair.Left == "product" && newIdx == f.Index {continue}
                    newRefs = append(newRefs, newIdx)
                } else {
                    newRefs = append(newRefs, ref)}}
            newVal := strings.Join(newRefs, " ")
            if newVal != pair.Right {
                f.Pairs[i].Right = newVal
                changed = true}
		case "content":
            content := pair.Right
            // Переписываем трансклюзии в обычном контенте
            ctPair := ""
            for _, p := range f.Pairs {
                if p.Left == "content-type" {ctPair = p.Right; break}}
            if ctPair == "json" {
                // JSON: ключи — индексы, переписываем их
                var obj map[string]string
                if json.Unmarshal([]byte(content), &obj) == nil {
                    newObj := make(map[string]string, len(obj))
                    for k, v := range obj {
                        newK := k
                        if newIdx, ok := indexMap[k]; ok {newK = newIdx}
                        newObj[newK] = v}
                    if b, err := json.Marshal(newObj); err == nil {
                        content = string(b)}}} else {
                for oldIdx, newIdx := range indexMap {
                    content = strings.ReplaceAll(content, "{{"+oldIdx+"}}", "{{"+newIdx+"}}")
                    content = strings.ReplaceAll(content, "[["+oldIdx+"]]", "[["+newIdx+"]]")}}
            if content != pair.Right {
                f.Pairs[i].Right = content
                changed = true}}}
	if changed {_ = WriteFrame(dstProjectID, f)}}
return len(cloned), nil }
// CollectImportFrames собирает фреймы для импорта в два прохода.
// Проход 1 — центральный узел и его связи первого уровня.
// Проход 2 — связи всех собранных фреймов (второй уровень).
func CollectImportFrames(projectID string, indices []string) ([]Frame, error) {
p, err := LoadProject(projectID)
if err != nil {return nil, err}
seen   := make(map[string]bool)
result := []Frame{}
// add добавляет фрейм в result если его титул ещё не встречался
add := func(f Frame) bool {
	title := FramePairRight(f, "title")
	if seen[title] {return false}
	seen[title] = true
	result = append(result, f)
	return true}
// addByIndex загружает фрейм по индексу и добавляет его
addByIndex := func(idx string) {
	if f, err := LoadFrame(projectID, idx); err == nil {add(f)}}
// addProductNeighbors добавляет фреймы из show-product через JSON контент
// Для каждого продукционного фрейма из слота product читаем его JSON контент
// и добавляем те фреймы из show-product которые есть как ключи в JSON
addProductNeighbors := func(f Frame) {
	showProduct := ParseTokenList(FramePairRight(f, "show-product"))
	if len(showProduct) == 0 {return}
	for _, prodIdx := range ParseTokenList(FramePairRight(f, "product")) {
		prod, err2 := LoadFrame(projectID, prodIdx)
		if err2 != nil {continue}
		var jsonContent map[string]string
		if json.Unmarshal([]byte(FramePairRight(prod, "content")), &jsonContent) != nil {continue}
		for _, key := range showProduct {
            if _, ok := jsonContent[key]; ok {
                addByIndex(key)}}}}
// addFrameLinks добавляет все связи фрейма — из его слотов и из мета-фрейма
addFrameLinks := func(f Frame) {
	// Связи из слотов фрейма
	for _, slot := range []string{"alias", "tags", "links", "includes", "product"} {
		for _, ref := range ParseTokenList(FramePairRight(f, slot)) {
			addByIndex(ref)}}
	// Связи через JSON продукций
	addProductNeighbors(f)
	// Входящие связи из мета-фрейма
	for _, slot := range []string{"tags", "links", "includes", "product"} {
		m := parseMapSlot(p.Meta.Frame, slot, p.Meta.FileOrderLen)
		for _, ref := range m[f.Index] {
			addByIndex(ref)}}}
// Проход 1 — центральный узел и его связи первого уровня
for _, idx := range indices {
	f, err := LoadFrame(projectID, idx)
	if err != nil {continue}
	if !add(f) {continue}
	addFrameLinks(f)}
// Проход 2 — связи всех собранных фреймов (второй уровень)
// Фиксируем длину result после прохода 1 — итерируем только по ним
firstLevel := len(result)
for i := 0; i < firstLevel; i++ {
	addFrameLinks(result[i])}
return result, nil }
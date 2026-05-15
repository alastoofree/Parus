//parus/unison/index.go

package unison

import (
	"fmt"
	"strings")

var frameSystemSlots = map[string]bool{"id": true, "modified": true, "title": true, "alias": true, 
"tags": true, "content": true, "content-type": true, "links": true, "includes": true, "icon": true, 
"color": true, "rank": true, "tab": true, "i-sort": true, "west": true, "nord": true, 
"graph": true, "list": true, "match-nord": true, "match-south": true, "chrono": true, 
"i-chrono": true, "dict": true, "i-dict": true, "show-product": true, "product": true,}

func BuildIndex(projectID, frameIndex string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
frame, err := LoadFrame(projectID, frameIndex)
if err != nil {return fmt.Errorf("load frame: %w", err)}
if FramePairRight(frame, "content-type") == "json" {return nil}
titleMap     := TitleMap(p.Meta)
titleToIndex := make(map[string]string, len(titleMap))
for idx, title := range titleMap {
	titleToIndex[strings.ToLower(title)] = idx}
meta := p.Meta
aTags     := parseMapSlot(meta.Frame, "tags",     p.Meta.FileOrderLen)
aLinks    := parseMapSlot(meta.Frame, "links",    p.Meta.FileOrderLen)
aIncludes := parseMapSlot(meta.Frame, "includes", p.Meta.FileOrderLen)
aProduct  := parseMapSlot(meta.Frame, "product",  p.Meta.FileOrderLen)
// Удаляем старые записи для frameIndex
for idx := range aTags     { aTags[idx]     = RemoveVal(aTags[idx],     frameIndex) }
for idx := range aLinks    { aLinks[idx]    = RemoveVal(aLinks[idx],    frameIndex) }
for idx := range aIncludes { aIncludes[idx] = RemoveVal(aIncludes[idx], frameIndex) }
for idx := range aProduct  { aProduct[idx]  = RemoveVal(aProduct[idx],  frameIndex) }
// Добавляем новые записи
for _, ref := range ParseTokenList(FramePairRight(frame, FrameSlotTags)) {
	if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != frameIndex {
		aTags[tgt] = AppendUnique(aTags[tgt], frameIndex)}}
for _, ref := range ParseTokenList(FramePairRight(frame, FrameSlotLinks)) {
	if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != frameIndex {
		aLinks[tgt] = AppendUnique(aLinks[tgt], frameIndex)}}
for _, ref := range ParseTokenList(FramePairRight(frame, FrameSlotIncludes)) {
	if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != frameIndex {
		aIncludes[tgt] = AppendUnique(aIncludes[tgt], frameIndex)}}
for _, ref := range ParseTokenList(FramePairRight(frame, "show-product")) {
	if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != frameIndex {
		aProduct[tgt] = AppendUnique(aProduct[tgt], frameIndex)}}
aChrono := ParseTokenList(FramePairRight(meta.Frame, "chrono"))
aChrono = RemoveVal(aChrono, frameIndex)
if FramePairRight(frame, "chrono") != "" {
	aChrono = AppendUnique(aChrono, frameIndex)}
allTitles := ParseTokenList(FramePairRight(p.Meta.Frame, "all-titles"))
updates := map[string]string{
	"chrono":   strings.Join(aChrono, " "),
	"tags":     fmtMap(aTags, allTitles, p.Meta.FileOrderLen),
	"links":    fmtMap(aLinks, allTitles, p.Meta.FileOrderLen),
	"includes": fmtMap(aIncludes, allTitles, p.Meta.FileOrderLen),
	"product":  fmtMap(aProduct, allTitles, p.Meta.FileOrderLen),}
updateMetaSlots(&meta, updates)
if p2, e2 := LoadProject(projectID); e2 == nil {
	for _, pair := range p2.Meta.Frame.Pairs {
		if pair.Left == "counts" {
			for i, mp := range meta.Frame.Pairs {
				if mp.Left == "counts" {meta.Frame.Pairs[i].Right = pair.Right; break}}; break}}}
if err := WriteMetaFrame(projectID, meta); err != nil {return err}
if projectID == MetaBaseName {return BuildGlobalIndex()}
return nil }
// 
func RebuildIndex(projectID string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {return fmt.Errorf("list frames: %w", err)}
titleMap     := TitleMap(p.Meta)
titleToIndex := make(map[string]string, len(titleMap))
for idx, title := range titleMap {
    titleToIndex[strings.ToLower(title)] = idx}
allTitles := ParseTokenList(FramePairRight(p.Meta.Frame, "all-titles"))
meta := p.Meta
aTags     := make(map[string][]string)
aLinks    := make(map[string][]string)
aIncludes := make(map[string][]string)
aProduct  := make(map[string][]string)
aChrono   := []string{}
for _, f := range frames {
    if FramePairRight(f, "content-type") == "json" {continue}
    idx := f.Index
    for _, ref := range ParseTokenList(FramePairRight(f, FrameSlotTags)) {
        if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != idx {
            aTags[tgt] = AppendUnique(aTags[tgt], idx)}}
    for _, ref := range ParseTokenList(FramePairRight(f, FrameSlotLinks)) {
        if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != idx {
            aLinks[tgt] = AppendUnique(aLinks[tgt], idx)}}
    for _, ref := range ParseTokenList(FramePairRight(f, FrameSlotIncludes)) {
        if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != idx {
            aIncludes[tgt] = AppendUnique(aIncludes[tgt], idx)}}
    for _, ref := range ParseTokenList(FramePairRight(f, "show-product")) {
        if tgt, ok := resolveRef(ref, titleToIndex); ok && tgt != idx {
            aProduct[tgt] = AppendUnique(aProduct[tgt], idx)}}
    if FramePairRight(f, "chrono") != "" {
        aChrono = AppendUnique(aChrono, idx)}}
updates := map[string]string{
	"chrono":   strings.Join(aChrono, " "),
	"tags":     fmtMap(aTags, allTitles, p.Meta.FileOrderLen),
	"links":    fmtMap(aLinks, allTitles, p.Meta.FileOrderLen),
	"includes": fmtMap(aIncludes, allTitles, p.Meta.FileOrderLen),
	"product":  fmtMap(aProduct, allTitles, p.Meta.FileOrderLen),}
updateMetaSlots(&meta, updates)
if err := WriteMetaFrame(projectID, meta); err != nil {return err}
if projectID == MetaBaseName {return BuildGlobalIndex()}
return nil}
// Count обновляет счётчики фреймов и тегов в мета-фрейме
func Count(projectID string, deltaFrames, deltaTags int) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
counts := FramePairRight(p.Meta.Frame, "counts")
parts := strings.Fields(counts)
nFrames, nTags := 0, 0
if len(parts) >= 2 {
	fmt.Sscanf(parts[0], "%d", &nFrames)
	fmt.Sscanf(parts[1], "%d", &nTags)}
nFrames += deltaFrames
nTags   += deltaTags
if nFrames < 0 {nFrames = 0}
if nTags < 0   {nTags = 0}
newVal := fmt.Sprintf("%d %d", nFrames, nTags)
for i, pair := range p.Meta.Frame.Pairs {
	if pair.Left == "counts" {
		p.Meta.Frame.Pairs[i].Right = newVal; break}}
return WriteMetaFrame(projectID, p.Meta) }
// RebuildTitles восстанавливает all-titles из файлов фреймов
func RebuildTitles(projectID string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {return fmt.Errorf("list frames: %w", err)}
tokens := make([]string, 0)
for _, f := range frames {
	pos := GaussOrder(f.Index)
	title := FramePairRight(f, "title")
	for len(tokens) <= pos {tokens = append(tokens, "")}
	tokens[pos] = title}
var sb strings.Builder
for i, t := range tokens {
	if i > 0 {sb.WriteString(" ")}
	sb.WriteString(quoteIfSpace(t))}
for i, pair := range p.Meta.Frame.Pairs {
	if pair.Left == "all-titles" {
		p.Meta.Frame.Pairs[i].Right = sb.String(); break}}
return WriteMetaFrame(projectID, p.Meta) }
// buildMetaIndex строит индексы для метабазы — теги, базы, совпадения
func buildMetaIndex(gm GlobalMeta) error {
meta, err := LoadProject(MetaBaseName)
if err != nil {return fmt.Errorf("load meta: %w", err)}
frames, err := ListFrames(MetaBaseName, meta.Meta)
if err != nil {return fmt.Errorf("list frames: %w", err)}
deepSouth  := parseIntDefault(FramePairRight(meta.Meta.Frame, "deep-south"), 3)
basesID     := ParseTokenList(FramePairRight(meta.Meta.Frame, "bases-id"))
basesTitles := ParseTokenList(FramePairRight(meta.Meta.Frame, "bases-title"))
// match-south — нечёткое совпадение титула фрейма мета-базы с титулом баз
matchSouth := make(map[string][]string)
for _, f := range frames {
	fTokens := metaTokenize(FramePairRight(f, "title"))
	for i, baseID := range basesID {
		if i >= len(basesTitles) {break}
		if metaTitlesOverlap(fTokens, metaTokenize(basesTitles[i]), deepSouth) {
			matchSouth[f.Index] = AppendUnique(matchSouth[f.Index], baseID)}}}
// Записываем match-nord и match-south в фреймы мета-базы
for _, f := range frames {
	nord := strings.Join(gm.MatchNord[f.Index], " ")
	south := strings.Join(matchSouth[f.Index], " ")
	changed := false
	for i, p := range f.Pairs {
		if p.Left == "match-nord" && p.Right != nord {
			f.Pairs[i].Right = nord; changed = true}
		if p.Left == "match-south" && p.Right != south {
			f.Pairs[i].Right = south; changed = true}}
	if changed {_ = WriteFrame(MetaBaseName, f)}}
// Внутренние теги мета-базы — связи между фреймами через слот tags
titleToIdx  := make(map[string]string)
tagFrameMap := make(map[string]Frame)
for _, f := range frames {
	t := FramePairRight(f, "title")
	if t != "" {titleToIdx[strings.ToLower(t)] = f.Index}
	for _, a := range ParseTokenList(FramePairRight(f, "alias")) {
		if a != "" {titleToIdx[strings.ToLower(a)] = f.Index}}
	tagFrameMap[f.Index] = f}
metaTagsOf := make(map[string][]string)
for _, f := range frames {
	src := f.Index
	for _, ref := range ParseTokenList(FramePairRight(f, "tags")) {
		tgt := ""
		if _, ok := tagFrameMap[ref]; ok {tgt = ref
		} else if idx, ok := titleToIdx[strings.ToLower(ref)]; ok {tgt = idx}
		if tgt != "" && tgt != src {
			metaTagsOf[src] = AppendUnique(metaTagsOf[src], tgt)}}}
for _, f := range frames {
	newTags := strings.Join(metaTagsOf[f.Index], " ")
	if FramePairRight(f, "tags") == newTags {continue}
	found := false
	for i, p := range f.Pairs {
		if p.Left == "tags" {f.Pairs[i].Right = newTags; found = true; break}}
	if !found {
		f.Pairs = append(f.Pairs, SlotPair{
			TypeIndex: ToGauss(len(f.Pairs), MetaBaseSlotLen),
			Left:      "tags", Right: newTags})}
	_ = WriteFrame(MetaBaseName, f)}
    // Вычисляем w-counts для каждого фрейма мета-базы
    aTags     := parseMapSlot(meta.Meta.Frame, "tags",     MetaBaseFileLen)
    aLinks    := parseMapSlot(meta.Meta.Frame, "links",    MetaBaseFileLen)
    aIncludes := parseMapSlot(meta.Meta.Frame, "includes", MetaBaseFileLen)
    aProduct  := parseMapSlot(meta.Meta.Frame, "product",  MetaBaseFileLen)
    allTitles := ParseTokenList(FramePairRight(meta.Meta.Frame, "all-titles"))
    var wParts []string
    for i := range allTitles {
    gauss := ToGauss(i, MetaBaseFileLen)
    f, ok := tagFrameMap[gauss]
        if !ok {wParts = append(wParts, "[ 0 0 0 0 0 0 0 0 0 0 0 0 0 ]"); continue}
        na  := len(ParseTokenList(FramePairRight(f, "alias")))
        nlt := len(aTags[gauss])
        nrt := len(ParseTokenList(FramePairRight(f, "tags")))
        nll := len(aLinks[gauss])
        nrl := len(ParseTokenList(FramePairRight(f, "links")))
        nli := len(aIncludes[gauss])
        nri := len(ParseTokenList(FramePairRight(f, "includes")))
        nlp := len(aProduct[gauss])
        nrp := len(ParseTokenList(FramePairRight(f, "product")))
        nbn := len(ParseTokenList(FramePairRight(f, "match-nord")))
        nbs := len(ParseTokenList(FramePairRight(f, "match-south")))
        wParts = append(wParts, fmt.Sprintf("[ %d %d %d %d %d %d %d %d %d %d %d 0 0 ]",
            na, nlt, nrt, nll, nrl, nli, nri, nlp, nrp, nbn, nbs))}
    gm.WCounts = strings.Join(wParts, " ")
return writeGlobalMeta(gm) }
// 
func RebuildTagCount(projectID string) error {
p, err := LoadProject(projectID)
if err != nil {return err}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {return err}
nTags := 0
for _, f := range frames {
    if r := FramePairRight(f, "rank"); r != "" {nTags++}}
nFrames := 0
for _, f := range frames {
    if FramePairRight(f, "content-type") != "json" {nFrames++}}
newVal := fmt.Sprintf("%d %d", nFrames, nTags)
for i, pair := range p.Meta.Frame.Pairs {
    if pair.Left == "counts" {
        p.Meta.Frame.Pairs[i].Right = newVal; break}}
return WriteMetaFrame(projectID, p.Meta) }
// parus/unison/global-meta.go — метафрейм метабазы и построение индекса

package unison

import (
	"fmt"
	"os"
	"path/filepath"
	"strings")

const (
	MetaBaseName    = "00000-META-BASE"
	MetaBaseFileLen = 5
	MetaBaseSlotLen = 1)
// GlobalMeta — данные мета-фрейма мета-базы
type GlobalMeta struct {
	BasesID     []string
	BasesTitles []string
	BasesCounts []string
	Counts      string
	GTags       string
	GLinks      string
	GIncludes   string
	GProduct    string
    WCounts     string
	MatchNord   map[string][]string}
// GlobalFrame — фрейм мета-базы
type GlobalFrame struct {
	Index   string
	Title   string
	Rank    string
	Aliases []string}
// SearchResult — результат глобального поиска
type SearchResult struct {
	BaseID     string `json:"baseId"`
	BaseTitle  string `json:"baseTitle"`
	FrameIndex string `json:"frameIndex"`
	FrameTitle string `json:"frameTitle"`
	TagIndex   string `json:"tagIndex"`
	TagTitle   string `json:"tagTitle"`}
// InitMetaBase создаёт мета-базу если не существует
func InitMetaBase() error {
dir := filepath.Join(DataDir, MetaBaseName)
if _, err := os.Stat(dir); err == nil {return nil}
if err := os.MkdirAll(dir, 0755); err != nil {
	return fmt.Errorf("mkdir %s: %w", MetaBaseName, err)}
if err := os.MkdirAll(filepath.Join(dir, "img"), 0755); err != nil {
	return fmt.Errorf("mkdir %s/img: %w", MetaBaseName, err)}
return writeGlobalMeta(GlobalMeta{}) }
// writeGlobalMeta записывает мета-фрейм мета-базы
func writeGlobalMeta(gm GlobalMeta) error {
existing := map[string]string{}
path := filepath.Join(DataDir, MetaBaseName, globalMetaFileName())
if frame, err := ParseMetaFrame(path); err == nil {
	for _, p := range frame.Frame.Pairs {existing[p.Left] = p.Right}}
// 
get := func(key, def string) string {
	if v, ok := existing[key]; ok && v != "" {return v}
	return def}
pairs := []SlotPair{
	{TypeIndex: "n", Left: "id",           Right: get("id", NewULID())},
	{TypeIndex: "m", Left: "",             Right: ""},
	{TypeIndex: "l", Left: "a-sort",       Right: get("a-sort", "")},
	{TypeIndex: "r", Left: "o-lens",       Right: get("o-lens", "")},
	{TypeIndex: "h", Left: "ranks",        Right: get("ranks", "")},
	{TypeIndex: "g", Left: "tabs",         Right: get("tabs", "")},
	{TypeIndex: "k", Left: "west",         Right: get("west", "")},
	{TypeIndex: "c", Left: "g-tags",       Right: gm.GTags},
	{TypeIndex: "s", Left: "deep-nord",    Right: get("deep-nord", "")},
	{TypeIndex: "z", Left: "deep-south",   Right: get("deep-south", "")},
	{TypeIndex: "t", Left: "graph",        Right: get("graph", "")},
	{TypeIndex: "d", Left: "list",         Right: get("list", "")},
	{TypeIndex: "b", Left: "chrono",       Right: get("chrono", "")},
	{TypeIndex: "p", Left: "g-links",      Right: gm.GLinks},
	{TypeIndex: "f", Left: "g-includes",   Right: gm.GIncludes},
	{TypeIndex: "v", Left: "counts",       Right: gm.Counts},
    {TypeIndex: "w", Left: "w-counts",     Right: gm.WCounts},
	{TypeIndex: "u", Left: "bases-counts", Right: joinBracketed(gm.BasesCounts)},
	{TypeIndex: "o", Left: "bases-id",     Right: strings.Join(gm.BasesID, " ")},
	{TypeIndex: "a", Left: "bases-title",  Right: joinQuoted(gm.BasesTitles)},
	{TypeIndex: "e", Left: "g-product",    Right: gm.GProduct},
	{TypeIndex: "i", Left: "all-titles",   Right: get("all-titles", "")},
	{TypeIndex: "y", Left: "tags",         Right: get("tags", "")},
	{TypeIndex: "j", Left: "links",        Right: get("links", "")},
	{TypeIndex: "q", Left: "includes",     Right: get("includes", "")},
	{TypeIndex: "x", Left: "product",      Right: get("product", "")},}
meta := MetaFrame{
	SlotOrderLen: MetaBaseSlotLen,
	FileOrderLen: MetaBaseFileLen,
	Frame:        Frame{Index: strings.Repeat("0", MetaBaseFileLen), Pairs: pairs},}
return WriteMetaFrame(MetaBaseName, meta) }
// globalMetaFileName возвращает имя файла мета-фрейма
func globalMetaFileName() string {
return strings.Repeat("0", MetaBaseFileLen) + ".uic"}
// 
func collectBaseLinks(projectID string, globalByTitle map[string]string,
	allGlobalFrames []GlobalFrame, metaTokensMap map[string][]string, deepNord int,
	gTags, gLinks, gIncludes, gProduct, matchNord map[string][]string) error {
p, err := LoadProject(projectID)
if err != nil {return fmt.Errorf("load project: %w", err)}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {return fmt.Errorf("list frames: %w", err)}
for _, f := range frames {
	title   := strings.ToLower(FramePairRight(f, "title"))
	fTokens := metaTokenize(FramePairRight(f, "title"))
	if r := FramePairRight(f, "rank"); r != "" {
		if globalIdx, ok := globalByTitle[title]; ok {
			gTags[globalIdx] = AppendUnique(gTags[globalIdx], projectID)
			globalTitle := ""
			for _, gt := range allGlobalFrames {
				if gt.Index == globalIdx {globalTitle = gt.Title; break}}
			link := "[" + globalTitle + "](frame.html?id=" + MetaBaseName + "&frame=" + globalIdx + ")"
			currentContent := FramePairRight(f, "content")
			if !strings.Contains(currentContent, link) {
				newContent := currentContent
				if currentContent != "" {newContent = currentContent + "\n\n---\n\n" + link
				} else {newContent = link}
				for i, pair := range f.Pairs {
					if pair.Left == "content" {
						f.Pairs[i].Right = newContent
						_ = WriteFrame(projectID, f); break}}}}}
	for _, ref := range ParseTokenList(FramePairRight(f, "links")) {
		if globalIdx, ok := globalByTitle[strings.ToLower(ref)]; ok {
			gLinks[globalIdx] = AppendUnique(gLinks[globalIdx], projectID)}}
	for _, ref := range ParseTokenList(FramePairRight(f, "includes")) {
		if globalIdx, ok := globalByTitle[strings.ToLower(ref)]; ok {
			gIncludes[globalIdx] = AppendUnique(gIncludes[globalIdx], projectID)}}
	for _, ref := range ParseTokenList(FramePairRight(f, "show-product")) {
		if globalIdx, ok := globalByTitle[strings.ToLower(ref)]; ok {
			gProduct[globalIdx] = AppendUnique(gProduct[globalIdx], projectID)}}
	for mfIdx, mfTokens := range metaTokensMap {
		if metaTitlesOverlap(fTokens, mfTokens, deepNord) {
			matchNord[mfIdx] = AppendUnique(matchNord[mfIdx], projectID)}}}
return nil }
// BuildGlobalIndex обходит все базы и обновляет g-* индексы мета-фрейма мета-базы
func BuildGlobalIndex() error {
projects, err := ListProjects()
if err != nil {return fmt.Errorf("list projects: %w", err)}
allGlobalFrames, _ := ListMetaFrames()
globalByTitle := make(map[string]string)
for _, gt := range allGlobalFrames {
	globalByTitle[strings.ToLower(gt.Title)] = gt.Index
	for _, a := range gt.Aliases {
		globalByTitle[strings.ToLower(a)] = gt.Index}}
gm           := GlobalMeta{}
gTags        := make(map[string][]string)
gLinks       := make(map[string][]string)
gIncludes    := make(map[string][]string)
gProduct     := make(map[string][]string)
matchNord    := make(map[string][]string)
// Читаем deep-nord из метафрейма мета-базы
metaProject, err := LoadProject(MetaBaseName)
if err != nil {return fmt.Errorf("load meta: %w", err)}
deepNord  := parseIntDefault(FramePairRight(metaProject.Meta.Frame, "deep-nord"),  3)
// Список фреймов мета-базы для построения match-nord
metaFrames, _ := ListFrames(MetaBaseName, metaProject.Meta)
metaTokensMap  := make(map[string][]string)
for _, mf := range metaFrames {
	metaTokensMap[mf.Index] = metaTokenize(FramePairRight(mf, "title"))}
for _, p := range projects {
	if p.ID == MetaBaseName {continue}
	gm.BasesID     = append(gm.BasesID,     p.ID)
	gm.BasesTitles = append(gm.BasesTitles, projectTitle(p))
	counts := ""
	for _, pair := range p.Meta.Frame.Pairs {
		if pair.Left == "counts" {counts = pair.Right; break}}
	gm.BasesCounts = append(gm.BasesCounts, counts)}
for _, baseID := range ParseTokenList(FramePairRight(metaProject.Meta.Frame, "bases-id")) {
	_ = collectBaseLinks(baseID, globalByTitle, allGlobalFrames, metaTokensMap, deepNord, gTags, gLinks, gIncludes, gProduct, matchNord)}
metaFrameCount := 0
for _, f := range metaFrames {
    if f.Index != strings.Repeat("0", MetaBaseFileLen) {metaFrameCount++}}
gm.Counts = fmt.Sprintf("%d %d", metaFrameCount, len(allGlobalFrames))
gm.GTags     = fmtGlobal(gTags, allGlobalFrames)
gm.GLinks    = fmtGlobal(gLinks, allGlobalFrames)
gm.GIncludes = fmtGlobal(gIncludes, allGlobalFrames)
gm.GProduct  = fmtGlobal(gProduct, allGlobalFrames)
gm.MatchNord = matchNord
return buildMetaIndex(gm) }
// BuildGlobalIndexForBase обновляет g-* индексы только для одной базы
func BuildGlobalIndexForBase(projectID string) error {
if projectID == MetaBaseName {return BuildGlobalIndex()}
metaProject, err := LoadProject(MetaBaseName)
if err != nil {return fmt.Errorf("load meta: %w", err)}
allGlobalFrames, _ := ListMetaFrames()
globalByTitle := make(map[string]string)
for _, gt := range allGlobalFrames {
    globalByTitle[strings.ToLower(gt.Title)] = gt.Index
    for _, a := range gt.Aliases {
        globalByTitle[strings.ToLower(a)] = gt.Index}}
// Читаем существующие g-* из мета-фрейма
gTags     := parseGlobalSlot(FramePairRight(metaProject.Meta.Frame, "g-tags"))
gLinks    := parseGlobalSlot(FramePairRight(metaProject.Meta.Frame, "g-links"))
gIncludes := parseGlobalSlot(FramePairRight(metaProject.Meta.Frame, "g-includes"))
gProduct  := parseGlobalSlot(FramePairRight(metaProject.Meta.Frame, "g-product"))
// Удаляем старые записи для данной базы
for idx := range gTags     {gTags[idx]     = RemoveVal(gTags[idx],     projectID)}
for idx := range gLinks    {gLinks[idx]    = RemoveVal(gLinks[idx],    projectID)}
for idx := range gIncludes {gIncludes[idx] = RemoveVal(gIncludes[idx], projectID)}
for idx := range gProduct  {gProduct[idx]  = RemoveVal(gProduct[idx],  projectID)}
// Обходим фреймы базы
deepNord := parseIntDefault(FramePairRight(metaProject.Meta.Frame, "deep-nord"), 3)
metaFrames, _ := ListFrames(MetaBaseName, metaProject.Meta)
metaTokensMap := make(map[string][]string)
for _, mf := range metaFrames {
    metaTokensMap[mf.Index] = metaTokenize(FramePairRight(mf, "title"))}
// Читаем существующий matchNord — сохраняем записи других баз
matchNord := make(map[string][]string)
for _, mf := range metaFrames {
    for _, baseID := range ParseTokenList(FramePairRight(mf, "match-nord")) {
        if baseID != projectID {
            matchNord[mf.Index] = AppendUnique(matchNord[mf.Index], baseID)}}}
basesID := ParseTokenList(FramePairRight(metaProject.Meta.Frame, "bases-id"))
if !contains(basesID, projectID) {
	updateMetaSlots(&metaProject.Meta, map[string]string{
		"bases-id": strings.Join(append(basesID, projectID), " ")})
	_ = WriteMetaFrame(MetaBaseName, metaProject.Meta)}
_ = collectBaseLinks(projectID, globalByTitle, allGlobalFrames, metaTokensMap, deepNord, gTags, gLinks, gIncludes, gProduct, matchNord)
// Форматируем g-* слоты. Записываем в мета-фрейм
updateMetaSlots(&metaProject.Meta, map[string]string{
	"g-tags":     fmtGlobal(gTags, allGlobalFrames),
	"g-links":    fmtGlobal(gLinks, allGlobalFrames),
	"g-includes": fmtGlobal(gIncludes, allGlobalFrames),
	"g-product":  fmtGlobal(gProduct, allGlobalFrames),})
if err := WriteMetaFrame(MetaBaseName, metaProject.Meta); err != nil {return err}
// Обновляем match-nord в фреймах мета-базы
gm := GlobalMeta{MatchNord: matchNord}
return buildMetaIndex(gm) }
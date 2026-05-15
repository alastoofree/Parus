// parus/unison/terms/term.go

package terms

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"parus/unison")

const TermBaseDir = "./KATER"
var termSlotKey = map[string]string{
	"n": "name",         "m": "alias",
	"l": "uri",          "r": "rank",
	"h": "tags",         "g": "links",
	"k": "color",        "c": "content",
	"s": "includes",     "z": "i-sort",
	"t": "tab",          "d": "dict",
	"b": "i-dict",       "p": "deep",
	"f": "graph",        "v": "list",
	"w": "west",         "u": "nord",
	"o": "chrono",       "a": "i-chrono",
	"e": "show-product", "i": "product",
	"y": "preposition",  "j": "attribute",
	"q": "prefix",       "x": "suffix",
	"0": "counts",       "1": "themes",
	"2": "reserv-1",     "3": "reserv-2",
	"4": "reserv-3",     "5": "reserv-4",
	"6": "reserv-5",     "7": "i-process",
	"8": "process",      "9": "q-btns",}
var termSlotLetter = map[string]string{}

func init() {
	for k, v := range termSlotKey {termSlotLetter[v] = k} }
// TermFrame — фрейм терминологической базы
type TermFrame struct {
	Path  string
	Pairs []unison.SlotPair}
// TermPairRight возвращает значение правого слота по имени левого
func TermPairRight(f TermFrame, key string) string {
for _, p := range f.Pairs {
	if p.Left == key {return p.Right}}
return "" }
// translit преобразует строку в латиницу для построения пути
var translitMap = map[rune]string{
	'а': "a",  'б': "b",  'в': "v",  'г': "g",  'д': "d",
	'е': "e",  'ё': "yo", 'ж': "zh", 'з': "z",  'и': "i",
	'й': "j",  'к': "k",  'л': "l",  'м': "m",  'н': "n",
	'о': "o",  'п': "p",  'р': "r",  'с': "s",  'т': "t",
	'у': "u",  'ф': "f",  'х': "h",  'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "sch",'ъ': "",   'ы': "y",  'ь': "",
	'э': "e",  'ю': "yu", 'я': "ya",
	'А': "A",  'Б': "B",  'В': "V",  'Г': "G",  'Д': "D",
	'Е': "E",  'Ё': "YO", 'Ж': "ZH", 'З': "Z",  'И': "I",
	'Й': "J",  'К': "K",  'Л': "L",  'М': "M",  'Н': "N",
	'О': "O",  'П': "P",  'Р': "R",  'С': "S",  'Т': "T",
	'У': "U",  'Ф': "F",  'Х': "H",  'Ц': "TS", 'Ч': "CH",
	'Ш': "SH", 'Щ': "SCH",'Ъ': "",   'Ы': "Y",  'Ь': "",
	'Э': "E",  'Ю': "YU", 'Я': "YA",}
func translit(s string) string {
var sb strings.Builder
for _, r := range s {
	if lat, ok := translitMap[r]; ok {
		sb.WriteString(lat)
	} else {
		sb.WriteRune(r)}}
return sb.String() }
// termPath строит путь к файлу термина из его имени
func termPath(name string) string {
lat := translit(strings.ToUpper(name))
lat = strings.ReplaceAll(lat, " ", "_")
parts := []string{TermBaseDir}
for _, r := range lat {
	parts = append(parts, string(r))}
dir := filepath.Join(parts...)
filename := strings.ToLower(translit(name)) + ".term"
return filepath.Join(dir, filename) }
// LoadTerm загружает .term файл
func LoadTerm(path string) (TermFrame, error) {
f, err := os.Open(path)
if err != nil {return TermFrame{}, fmt.Errorf("open %s: %w", path, err)}
defer f.Close()
leftMap  := make(map[string]string)
rightMap := make(map[string]string)
order    := []string{}
seen     := make(map[string]bool)
scanner := bufio.NewScanner(f)
scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
for scanner.Scan() {
	line := scanner.Text()
	if line == "" {continue}
	idx := strings.IndexByte(line, ' ')
	var key, value string
	if idx < 0 {key = line; value = ""
	} else {key = line[:idx]; value = line[idx+1:]}
	sk, ok := unison.ParseSlotKey(key)
	if !ok {continue}
	if !seen[sk.Type] {
		seen[sk.Type] = true
		order = append(order, sk.Type)}
	switch sk.Side {
	case "l": leftMap[sk.Type] = value
	case "r": rightMap[sk.Type] = value}}
if err := scanner.Err(); err != nil {
	return TermFrame{}, fmt.Errorf("scan %s: %w", path, err)}
pairs := make([]unison.SlotPair, 0, len(order))
for _, typeKey := range order {
	pairs = append(pairs, unison.SlotPair{
		TypeIndex: typeKey,
		Left:      leftMap[typeKey],
		Right:     strings.ReplaceAll(rightMap[typeKey], `\n`, "\n"),})}
return TermFrame{Path: path, Pairs: pairs}, nil }
// WriteTerm записывает .term файл
func WriteTerm(f TermFrame) error {
if err := os.MkdirAll(filepath.Dir(f.Path), 0755); err != nil {
	return fmt.Errorf("mkdir: %w", err)}
var sb strings.Builder
for _, pair := range f.Pairs {
	letter := termSlotLetter[pair.Left]
	if letter == "" {letter = pair.TypeIndex}
	fmt.Fprintf(&sb, "l-%s %s\n", letter, pair.Left)
	fmt.Fprintf(&sb, "r-%s %s\n", letter, strings.ReplaceAll(pair.Right, "\n", `\n`))}
return os.WriteFile(f.Path, []byte(sb.String()), 0644) }
// UniqueTermName возвращает уникальное имя терма
func UniqueTermName(name string) string {
base := name
counter := 2
for {
    p := termPath(base)
    if _, err := os.Stat(p); os.IsNotExist(err) {break}
    base = fmt.Sprintf("%s-%d", name, counter)
    counter++}
return base }
// CreateTerm создаёт новый .term файл
func CreateTerm(name string) (TermFrame, error) {
path := termPath(name)
name = UniqueTermName(name)
path = termPath(name)
f := TermFrame{
	Path: path,
	Pairs: []unison.SlotPair{
		{TypeIndex: "n", Left: "name",         Right: name},
		{TypeIndex: "m", Left: "alias",        Right: ""},
        {TypeIndex: "l", Left: "uri",          Right: filepath.Dir(path)},
		{TypeIndex: "r", Left: "rank",         Right: ""},
		{TypeIndex: "h", Left: "tags",         Right: ""},
		{TypeIndex: "g", Left: "links",        Right: ""},
		{TypeIndex: "k", Left: "color",        Right: ""},
		{TypeIndex: "c", Left: "content",      Right: ""},
		{TypeIndex: "s", Left: "includes",     Right: ""},
		{TypeIndex: "z", Left: "i-sort",       Right: ""},
		{TypeIndex: "t", Left: "tab",          Right: ""},
		{TypeIndex: "d", Left: "dict",         Right: ""},
		{TypeIndex: "b", Left: "i-dict",       Right: ""},
		{TypeIndex: "p", Left: "deep",         Right: ""},
		{TypeIndex: "f", Left: "graph",        Right: ""},
		{TypeIndex: "v", Left: "list",         Right: ""},
		{TypeIndex: "w", Left: "west",         Right: ""},
		{TypeIndex: "u", Left: "nord",         Right: ""},
		{TypeIndex: "o", Left: "chrono",       Right: ""},
		{TypeIndex: "a", Left: "i-chrono",     Right: ""},
		{TypeIndex: "e", Left: "show-product", Right: ""},
		{TypeIndex: "i", Left: "product",      Right: ""},
		{TypeIndex: "y", Left: "preposition",  Right: ""},
		{TypeIndex: "j", Left: "attribute",    Right: ""},
		{TypeIndex: "q", Left: "prefix",       Right: ""},
		{TypeIndex: "x", Left: "suffix",       Right: ""},
		{TypeIndex: "0", Left: "counts",       Right: ""},
		{TypeIndex: "1", Left: "themes",       Right: ""},
		{TypeIndex: "2", Left: "reserv-1",     Right: ""},
		{TypeIndex: "3", Left: "reserv-2",     Right: ""},
		{TypeIndex: "4", Left: "reserv-3",     Right: ""},
		{TypeIndex: "5", Left: "reserv-4",     Right: ""},
		{TypeIndex: "6", Left: "reserv-5",     Right: ""},
		{TypeIndex: "7", Left: "i-process",    Right: ""},
		{TypeIndex: "8", Left: "process",      Right: ""},
		{TypeIndex: "9", Left: "q-btns",       Right: ""},},}
return f, WriteTerm(f) }
// CloneTerm клонирует терм
func CloneTerm(sourcePath, newName, targetUri string) (TermFrame, error) {
dir := targetUri
if dir == "" {dir = filepath.Dir(sourcePath)}
filename := strings.ToLower(translit(newName)) + ".term"
path := filepath.Join(dir, filename)
if _, err := os.Stat(path); err == nil {
    return TermFrame{}, fmt.Errorf("term already exists: %s", newName)}
f := TermFrame{Path: path, Pairs: []unison.SlotPair{}}
return f, WriteTerm(f) }
// ChronoTermRow — строка хронологической таблицы термов
type ChronoTermRow struct {
    TermPath  string            `json:"termPath"`
    TermName  string            `json:"termName"`
    TermColor string            `json:"termColor"`
    DateVal   string            `json:"dateVal"`
    Extras    map[string]string `json:"extras"`}
// GetTermChrono возвращает хронологическую таблицу для терма-тега
func GetTermChrono(path string) ([]ChronoTermRow, error) {
f, err := LoadTerm(path)
if err != nil {return nil, fmt.Errorf("load term: %w", err)}
chronoVal  := TermPairRight(f, "chrono")
iChronoVal := TermPairRight(f, "i-chrono")
if chronoVal == "" {return []ChronoTermRow{}, nil}
retro     := strings.HasPrefix(chronoVal, "-")
extraCols := unison.ParseTokenList(iChronoVal)
allTerms, err := ListTerms()
nameMap := make(map[string]string)
for _, t := range allTerms {
    name := TermPairRight(t, "name")
    if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
//
resolveLinks := func(val string) string {
    return regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllStringFunc(val, func(match string) string {
        name := strings.TrimSpace(match[2 : len(match)-2])
        tp, ok := nameMap[strings.ToLower(name)]
        if !ok {return name}
        return `<a href="term.html?path=` + tp + `">` + name + `</a>`})}
if err != nil {return nil, fmt.Errorf("list terms: %w", err)}
var rows []ChronoTermRow
for _, t := range allTerms {
    sat, err2 := LoadTermSatellite(t.Path)
    if err2 != nil {continue}
    dateVal := sat["date"]
    if dateVal == "" {continue}
    extras := make(map[string]string)
    for _, col := range extraCols {
        if v, ok := sat[col]; ok {extras[col] = resolveLinks(v)}}
    rows = append(rows, ChronoTermRow{
        TermPath:  t.Path,
        TermName:  TermPairRight(t, "name"),
        TermColor: TermPairRight(t, "color"),
        DateVal:   dateVal,
        Extras:    extras,})}
if retro {
    for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {rows[i], rows[j] = rows[j], rows[i]}
} else {
    for i := 1; i < len(rows); i++ {
        for j := i; j > 0 && rows[j].DateVal < rows[j-1].DateVal; j-- {
            rows[j], rows[j-1] = rows[j-1], rows[j]}}}
return rows, nil }
// ProcessTermAlias — алиас с путём
type ProcessTermAlias struct {
    Name string `json:"name"`
    Path string `json:"path"`}
// ProcessTermRow — строка таблицы процесса
type ProcessTermRow struct {
    Key      string             `json:"key"`
    TermA    string             `json:"termA"`
    Prep     string             `json:"prep"`
    TermB    string             `json:"termB"`
    PathA    string             `json:"pathA"`
    PathP    string             `json:"pathP"`
    PathB    string             `json:"pathB"`
    AliasesA []ProcessTermAlias `json:"aliasesA"`
    AliasesP []ProcessTermAlias `json:"aliasesP"`
    AliasesB []ProcessTermAlias `json:"aliasesB"`}
// toGauss4 возвращает гаусс-индекс четвёртого порядка
func toGauss4(n int) string {
const abc = "oiztyvslanxedbhfmg"
r := make([]byte, 4)
for i := 3; i >= 0; i-- {
    r[i] = abc[n%18]
    n /= 18}
return string(r) }
// buildAliases строит список алиасов с путями для терма
func buildAliases(termName string, nameMap map[string]string) []ProcessTermAlias {
tp, ok := nameMap[strings.ToLower(termName)]
if !ok {return nil}
f, err := LoadTerm(tp)
if err != nil {return nil}
raw := TermPairRight(f, "alias")
if raw == "" {return nil}
var result []ProcessTermAlias
for _, a := range strings.Fields(raw) {
    ap, ok2 := nameMap[strings.ToLower(a)]
    if !ok2 {continue}
    result = append(result, ProcessTermAlias{
        Name: a,
        Path: strings.TrimPrefix(ap, "./")})}
return result }
// GetTermProcess возвращает таблицу процесса для терма
func GetTermProcess(path string) ([]ProcessTermRow, error) {
f, err := LoadTerm(path)
if err != nil {return nil, fmt.Errorf("load term: %w", err)}
processVal  := TermPairRight(f, "process")
iProcessVal := TermPairRight(f, "i-process")
if processVal == "" {return []ProcessTermRow{}, nil}
allTerms, err := ListTerms()
if err != nil {return nil, fmt.Errorf("list terms: %w", err)}
nameMap := make(map[string]string)
for _, t := range allTerms {
    name := TermPairRight(t, "name")
    if name == "" {continue}
    nameMap[strings.ToLower(name)] = t.Path}
satPaths := []string{path}
for _, pname := range unison.ParseTokenList(iProcessVal) {
    if tp, ok := nameMap[strings.ToLower(pname)]; ok {
        satPaths = append(satPaths, tp)}}
type sit struct{ key string; val string }
var allSits []sit
counter := 0
for _, sp := range satPaths {
    sat, err2 := LoadTermSatellite(sp)
    if err2 != nil {continue}
    localKeys := make([]string, 0, len(sat))
    for k := range sat {localKeys = append(localKeys, k)}
    for i := 1; i < len(localKeys); i++ {
        for j := i; j > 0 && unison.GaussOrder(localKeys[j]) < unison.GaussOrder(localKeys[j-1]); j-- {
            localKeys[j], localKeys[j-1] = localKeys[j-1], localKeys[j]}}
    for _, k := range localKeys {
        newKey := toGauss4(counter)
        counter++
        allSits = append(allSits, sit{key: newKey, val: sat[k]})}}
keys := make([]string, 0, len(allSits))
for _, s := range allSits {keys = append(keys, s.key)}
var rows []ProcessTermRow
sitMap := map[string]string{}
for _, s := range allSits {sitMap[s.key] = s.val}
for _, k := range keys {
    parts := strings.Fields(sitMap[k])
    if len(parts) != 3 {continue}
    nameA, nameP, nameB := parts[0], parts[1], parts[2]
    pathA, _ := nameMap[strings.ToLower(nameA)]
    pathP, _ := nameMap[strings.ToLower(nameP)]
    pathB, _ := nameMap[strings.ToLower(nameB)]
    rows = append(rows, ProcessTermRow{
        Key:      k,
        TermA:    nameA,
        Prep:     nameP,
        TermB:    nameB,
        PathA:    strings.TrimPrefix(pathA, "./"),
        PathP:    strings.TrimPrefix(pathP, "./"),
        PathB:    strings.TrimPrefix(pathB, "./"),
        AliasesA: buildAliases(nameA, nameMap),
        AliasesP: buildAliases(nameP, nameMap),
        AliasesB: buildAliases(nameB, nameMap)})}
return rows, nil }
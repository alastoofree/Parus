//parus/unison/parse.go

package unison

import (
	"crypto/rand"
	"fmt"
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time")    

// SlotKey — разобранный ключ строки .uic файла. Формат: [сторона]-[тип] или [индекс]
type SlotKey struct {
	Side  string // "l" или "r"
	Type  string
	Index string}
// SlotPair — пара строк: левый и правый слот. Продукция: обобщённая импликация
type SlotPair struct {
	TypeIndex string // гаусс-индекс типа
	Left      string
	Right     string}
// gaussChars — алфавит гаусс-кодирования
var gaussChars = []byte{'o', 'i', 'z', 't', 'y', 'v', 's', 'l', 'a', 'n', 'x', 'e', 'd', 'b', 'h', 'f', 'm', 'g'}
// ToGauss кодирует число n в гаусс-индекс заданной длины
func ToGauss(n, length int) string {
result := make([]byte, length)
for i := length - 1; i >= 0; i-- {
	result[i] = gaussChars[n%18]
	n /= 18}
return string(result) }
// GaussOrder возвращает числовой порядок гаусс-индекса для сортировки
func GaussOrder(s string) int {
n := 0
for _, c := range s {
	n = n*18 + strings.IndexRune("oiztyvslanxedbhfmg", c)}
return n }
// BuildCharOrder строит карту символ→вес из строки a-sort
func BuildCharOrder(aSort string) map[rune]int {
order := make(map[rune]int)
for _, c := range " -," {
	order[c] = -1}
pos := 0
for _, c := range strings.ToLower(aSort) {
	if c == ' ' {continue}
	order[c] = pos
	pos++}
return order }
// CompareByASort сравнивает два титула по карте порядка
func CompareByASort(a, b string, order map[rune]int) int {
ra := []rune(strings.ToLower(a))
rb := []rune(strings.ToLower(b))
maxLen := len(ra)
if len(rb) > maxLen {maxLen = len(rb)}
for i := 0; i < maxLen; i++ {
	if i >= len(ra) {return -1}
	if i >= len(rb) {return 1}
	oa, okA := order[ra[i]]
	ob, okB := order[rb[i]]
	if !okA {oa = 10000}
	if !okB {ob = 10000}
	if oa != ob {return oa - ob}}
return 0 }
// Алфавит ULID — 32 символа (Crockford Base32)
const ulidAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
// NewULID генерирует новый ULID [10 символов времени][16 случайных символов]
func NewULID() string {
// Время в миллисекундах
ms := time.Now().UnixMilli()
// Кодируем время (10 символов)
timeChars := make([]byte, 10)
for i := 9; i >= 0; i-- {
	timeChars[i] = ulidAlphabet[ms%32]
	ms /= 32}
// Случайные байты (10 байт -> 16 символов base32)
randBytes := make([]byte, 10)
_, err := rand.Read(randBytes)
if err != nil {
	panic(fmt.Sprintf("ulid: rand read failed: %v", err))}
// Кодируем случайную часть (16 символов)
randChars := make([]byte, 16)
// Используем 80 бит (10 байт) -> 16 * 5 бит
bits := uint64(0)
bitCount := 0
pos := 0
for _, b := range randBytes {
	bits = (bits << 8) | uint64(b)
	bitCount += 8
	for bitCount >= 5 {
		bitCount -= 5
		randChars[pos] = ulidAlphabet[(bits>>uint(bitCount))&0x1F]
		pos++}}
return string(timeChars) + string(randChars) }
// ValidULID проверяет что строка является валидным ULID
func ValidULID(s string) bool {
if len(s) != 26 {return false}
upper := strings.ToUpper(s)
for _, c := range upper {
	if !strings.ContainsRune(ulidAlphabet, c) {return false}}
return true }
// ParseSlotKey разбирает ключ строки .uic файла. Формат: [сторона]-[тип]
func ParseSlotKey(raw string) (SlotKey, bool) {
parts := strings.SplitN(raw, "-", 3)
if len(parts) < 2 {return SlotKey{}, false}
side := parts[0]
if side != "l" && side != "r" {return SlotKey{}, false}
typeKey := parts[1]
index := ""
if len(parts) == 3 {index = parts[2]}
return SlotKey{
	Side:  side,
	Type:  typeKey,
	Index: index,}, true }
// 
func parseMapSlot(f Frame, slotName string, fileOrderLen int) map[string][]string {
result := make(map[string][]string)
raw := FramePairRight(f, slotName)
if raw == "" { return result }
tokens := parseList(raw)
pos := 0
i   := 0
for i < len(tokens) {
	if tokens[i] == "[" {
		i++
		key := ToGauss(pos, fileOrderLen)
		for i < len(tokens) && tokens[i] != "]" {
			result[key] = append(result[key], tokens[i])
			i++}
		if i < len(tokens) {i++}
		pos++
	} else {i++}}
return result }
// 
func resolveRef(ref string, titleToIndex map[string]string) (string, bool) {
for _, idx := range titleToIndex {
	if idx == ref {return ref, true}}
return "", false }
// 
func parseIntDefault(s string, def int) int {
n := 0
for _, c := range s {
	if c < '0' || c > '9' { return def }
	n = n*10 + int(c-'0')}
if n == 0 { return def }
return n }
// ParseFrame читает .uic файл и возвращает Frame
func ParseFrame(path string) (Frame, error) {
f, err := os.Open(path)
if err != nil {
	return Frame{}, fmt.Errorf("open %s: %w", path, err)}
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
	if idx < 0 {
		key = line
		value = ""
	} else {
		key = line[:idx]
		value = line[idx+1:]}
	sk, ok := ParseSlotKey(key)
	if !ok {continue}
	if !seen[sk.Type] {
		seen[sk.Type] = true
		order = append(order, sk.Type)}
	switch sk.Side {
	case "l":
		leftMap[sk.Type] = value
	case "r":
		rightMap[sk.Type] = value}}
if err := scanner.Err(); err != nil {
	return Frame{}, fmt.Errorf("scan %s: %w", path, err)}
pairs := make([]SlotPair, 0, len(order))
for _, typeKey := range order {
	left := leftMap[typeKey]
	pairs = append(pairs, SlotPair{
		TypeIndex: typeKey,
		Left:      left,
		Right:     strings.ReplaceAll(rightMap[typeKey], `\n`, "\n"),})}
return Frame{Pairs: pairs}, nil }
// ParseMetaFrame читает метафрейм и возвращает MetaFrame. Имя файла: [порядок-файла].uic
// Пример: 000.uic -> FileOrderLen=3
func ParseMetaFrame(path string) (MetaFrame, error) {
name := filepath.Base(path)
name = strings.TrimSuffix(name, ".uic") 
if name == "" || strings.TrimLeft(name, "0") != "" {
	return MetaFrame{}, fmt.Errorf("invalid metaframe filename: %s", name)}
fileOrderLen := len(name)
frame, err := ParseFrame(path)
if err != nil {return MetaFrame{}, err}
frame.Index = strings.Repeat("0", fileOrderLen)
return MetaFrame{
	Frame:        frame,
	SlotOrderLen: 1,
	FileOrderLen: fileOrderLen}, nil }
// ParseTokenList разбирает строку токенов с поддержкой кавычек
func ParseTokenList(s string) []string {
var result []string
s = strings.TrimSpace(s)
for s != "" {
	var token string
	if len(s) > 0 && s[0] == '"' {
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			token = s[1:]
			s = ""
		} else {
			token = s[1 : end+1]
			s = strings.TrimSpace(s[end+2:])}
	} else {
		idx := strings.IndexByte(s, ' ')
		if idx < 0 {
			token = s
			s = ""
		} else {
			token = s[:idx]
			s = strings.TrimSpace(s[idx+1:])}}
	if token != "" {
		result = append(result, token)}}
return result }
// parseList — синоним ParseTokenList
func parseList(s string) []string { return ParseTokenList(s) }
// AppendUnique добавляет элемент в слайс если его там нет
func AppendUnique(slice []string, s string) []string {
for _, v := range slice {
	if v == s { return slice }}
return append(slice, s) }
// quoteIfSpace оборачивает строку в кавычки если содержит пробел
func quoteIfSpace(s string) string {
if strings.Contains(s, " ") { return `"` + s + `"` }
return s }
// 
func parseGlobalSlot(raw string) map[string][]string {
result := make(map[string][]string)
tokens := ParseTokenList(raw)
i := 0
for i < len(tokens) {
    idx := tokens[i]
    i++
    if i < len(tokens) && tokens[i] == "[" {
        i++
        for i < len(tokens) && tokens[i] != "]" {
            result[idx] = append(result[idx], tokens[i])
            i++}
        if i < len(tokens) {i++}}}
return result}
// 
func RemoveVal(slice []string, val string) []string {
result := slice[:0]
for _, v := range slice {
	if v != val {result = append(result, v)}}
return result }
// fmtMap форматирует map индекс→[]индекс в строку слота мета-фрейма
func fmtMap(m map[string][]string, allTitles []string, fileOrderLen int) string {
var b strings.Builder
for i := range allTitles {
	idx := ToGauss(i, fileOrderLen)
	b.WriteString(" [ ")
	for _, id := range m[idx] {b.WriteString(id + " ")}
	b.WriteString("] ")}
return strings.TrimSpace(b.String()) }
// fmtGlobal форматирует map индекс→[]baseID в строку g-* слота мета-фрейма
func fmtGlobal(m map[string][]string, frames []GlobalFrame) string {
var b strings.Builder
for _, gt := range frames {
	b.WriteString(gt.Index + " [ ")
	for _, baseID := range m[gt.Index] {b.WriteString(baseID + " ")}
	b.WriteString("] ")}
return strings.TrimSpace(b.String()) }
// extractBracketContent извлекает содержимое скобок для заданного index
// Формат слота: idx1 [ ref ref ... ] idx2 [ ref ... ]
func extractBracketContent(slot, index string) []string {
if slot == "" || index == "" { return nil }
// Ищем "index [...]" — используем FindAllStringSubmatch для надёжности
pattern := regexp.QuoteMeta(index) + `\s*\[([^\]]*)\]`
re, err := regexp.Compile(pattern)
if err != nil { return nil }
m := re.FindStringSubmatch(slot)
if len(m) < 2 { return nil }
content := strings.TrimSpace(m[1])
if content == "" { return nil }
return ParseTokenList(content) }
//
func stripContent(s string) string {
// Убираем HTML теги
s = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, " ")
// Убираем markdown символы
s = regexp.MustCompile(`[#*\[\]()` + "`" + `]`).ReplaceAllString(s, " ")
// Убираем лишние пробелы
s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
return strings.TrimSpace(s) }
// 
func regexpFind(pattern, s string) string {
re := regexp.MustCompile(pattern)
m  := re.FindStringSubmatch(s)
if len(m) < 2 { return "" }; return m[1] }
// projectTitle возвращает титул базы
func projectTitle(p Project) string {
for _, pair := range p.Meta.Frame.Pairs {
	if pair.Left == "title" {return pair.Right}}
return ""}
// joinQuoted оборачивает элементы с пробелами в кавычки
func joinQuoted(items []string) string {
var parts []string
for _, item := range items {
	if item == "" || strings.Contains(item, " ") {
		parts = append(parts, `"`+item+`"`)
	} else {parts = append(parts, item)}}
return strings.Join(parts, " ") }
// joinBracketed оборачивает элементы с пробелами в квадратные скобки
func joinBracketed(items []string) string {
var parts []string
for _, item := range items {
    parts = append(parts, "[ "+item+" ]")}
return strings.Join(parts, " ") }
// metaTokenize разбивает строку на токены по пробелу, дефису и подчёркиванию
func metaTokenize(s string) []string {
s = strings.ToLower(strings.TrimSpace(s))
return strings.FieldsFunc(s, func(r rune) bool {
	return r == ' ' || r == '-' || r == '_'}) }
// metaTitlesOverlap проверяет пересечение токенов двух строк по префиксу длиной match-south (depth)
func metaTitlesOverlap(a, b []string, depth int) bool {
for _, ta := range a {
	ra := []rune(ta)
	if len(ra) < depth {continue}
	for _, tb := range b {
		rb := []rune(tb)
		if len(rb) < depth {continue}
		d := depth
		if len(ra) < d {d = len(ra)}
		if len(rb) < d {d = len(rb)}
		if strings.HasPrefix(string(rb), string(ra[:d])) ||
			strings.HasPrefix(string(ra), string(rb[:d])) {return true}}}
return false }
// 
func contains(slice []string, val string) bool {
for _, v := range slice {if v == val {return true}}
return false }
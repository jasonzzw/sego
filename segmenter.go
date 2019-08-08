//Go中文分词
package sego

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	minTokenFrequency = 2 // 仅从字典文件中读取大于等于此频率的分词
)

// 分词器结构体
type Segmenter struct {
	dict   *Dictionary
	Phrase bool
}

// 该结构体用于记录Viterbi算法中某字元处的向前分词跳转信息
type jumper struct {
	minDistance float32
	token       *Token
}

// 返回分词器使用的词典
func (seg *Segmenter) Dictionary() *Dictionary {
	return seg.dict
}

// 从文件中载入词典
//
// 可以载入多个词典文件，文件名用","分隔，排在前面的词典优先载入分词，比如
// 	"用户词典.txt,通用词典.txt"
// 当一个分词既出现在用户词典也出现在通用词典中，则优先使用用户词典。
//
// 词典的格式为（每个分词一行）：
//	分词文本 频率 词性
func (seg *Segmenter) LoadDictionary(files string) {
	seg.dict = NewDictionary()

	for _, file := range strings.Split(files, ",") {
		log.Printf("loading sego dictionary %s", file)
		dictFile, err := os.Open(file)
		defer dictFile.Close()
		if err != nil {
			log.Fatalf("cannot load sego dictionary \"%s\" \n", file)
		}

		reader := bufio.NewReader(dictFile)
		var text string
		var freqText string
		var frequency int
		var pos string

		// 逐行读入分词

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			size, _ := fmt.Sscanf(line, "%s %s %s\n", &text, &freqText, &pos)
			if size == 0 {
				// 文件结束
				break
			} else if size < 2 {
				// 无效行
				continue
			} else if size == 2 {
				// 没有词性标注时设为空字符串
				pos = ""
			}

			// 解析词频
			frequency, err = strconv.Atoi(freqText)
			if err != nil {
				continue
			}

			// 过滤频率太小的词
			if frequency < minTokenFrequency {
				continue
			}

			// 将分词添加到字典中
			words := splitTextToWords([]byte(text), seg.Phrase)
			token := Token{text: words, frequency: frequency, pos: pos}
			seg.dict.addToken(token, seg.Phrase)
		}
	}

	// 计算每个分词的路径值，路径值含义见Token结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	log.Println("sego dictionary loading complete")
}

// For segmenting English Word
func (seg *Segmenter) LoadEnglishDictionary(files string) {
	seg.dict = NewDictionary()

	for _, file := range strings.Split(files, ",") {
		log.Printf("loading sego dictionary %s", file)
		dictFile, err := os.Open(file)
		defer dictFile.Close()
		if err != nil {
			log.Fatalf("cannot load sego dictionary \"%s\" \n", file)
		}

		reader := bufio.NewReader(dictFile)
		var text string
		var freqText string
		var frequency int
		var pos string

		// 逐行读入分词

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			size, _ := fmt.Sscanf(line, "%s %s %s\n", &text, &freqText, &pos)
			if size == 0 {
				// 文件结束
				break
			} else if size < 2 {
				// 无效行
				continue
			} else if size == 2 {
				// 没有词性标注时设为空字符串
				pos = ""
			}

			// 解析词频
			frequency, err = strconv.Atoi(freqText)
			if err != nil {
				continue
			}

			// 过滤频率太小的词
			if frequency < minTokenFrequency {
				continue
			}

			text = strings.ToLower(text)

			// 将分词添加到字典中
			words := splitEnglishTextToWords([]byte(text), seg.Phrase)
			token := Token{text: words, frequency: frequency, pos: pos}
			seg.dict.addToken(token, seg.Phrase)
		}
	}

	// 计算每个分词的路径值，路径值含义见Token结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	log.Println("sego dictionary loading complete")
}

func (seg *Segmenter) LoadPreLoadDictionary(preDict map[string]string) {
	seg.dict = NewDictionary()
	var text string
	var freqText string
	var pos string
	for cand, v := range preDict {
		size, _ := fmt.Sscanf(v, "%s %s %s\n", &text, &freqText, &pos)
		if size < 2 {
			// 无效行
			continue
		} else if size == 2 {
			// 没有词性标注时设为空字符串
			pos = ""
		}

		// 解析词频
		frequency, err := strconv.Atoi(freqText)
		if err != nil {
			continue
		}

		// 过滤频率太小的词
		if frequency < minTokenFrequency {
			continue
		}

		// 将分词添加到字典中
		words := splitTextToWords([]byte(cand), seg.Phrase)
		token := Token{text: words, frequency: frequency, pos: pos}
		seg.dict.addToken(token, seg.Phrase)
	}

	// 计算每个分词的路径值，路径值含义见Token结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	log.Println("sego dictionary loading complete")
}

// 对文本分词
//
// 输入参数：
//	bytes	UTF8文本的字节数组
//
// 输出：
//	[]Segment	划分的分词

func (seg *Segmenter) Segment(bytes []byte, joint string) []string {
	return seg.internalSegment(bytes, joint, false, "")
}

func (seg *Segmenter) SegmentExclude(bytes []byte, joint string, exclude string) []string {
	return seg.internalSegment(bytes, joint, false, exclude)
}

func (seg *Segmenter) SegmentEnglish(bytes []byte, joint string) []string {
	return seg.internalEnglishSegment(bytes, joint, false)
}

func (seg *Segmenter) internalEnglishSegment(bytes []byte, joint string, searchMode bool) []string {
	// 处理特殊情况
	if len(bytes) == 0 {
		return []string{}
	}
	// 划分字元
	text := splitEnglishTextToWords(bytes, seg.Phrase)
	return seg.segmentWords(text, joint, searchMode, "")
}

func (seg *Segmenter) internalSegment(bytes []byte, joint string, searchMode bool, exclude string) []string {
	// 处理特殊情况
	if len(bytes) == 0 {
		return []string{}
	}

	// 划分字元
	text := splitTextToWords(bytes, seg.Phrase)

	return seg.segmentWords(text, joint, searchMode, exclude)
}

func (seg *Segmenter) segmentWords(text []Text, joint string, searchMode bool, exclude string) []string {
	// 搜索模式下该分词已无继续划分可能的情况
	if searchMode && len(text) == 1 {
		return []string{}
	}

	// jumpers定义了每个字元处的向前跳转信息，包括这个跳转对应的分词，
	// 以及从文本段开始到该字元的最短路径值
	jumpers := make([]jumper, len(text))

	tokens := make([]*Token, seg.dict.maxTokenLength)
	for current := 0; current < len(text); current++ {
		// 找到前一个字元处的最短路径，以便计算后续路径值
		var baseDistance float32
		if current == 0 {
			// 当本字元在文本首部时，基础距离应该是零
			baseDistance = 0
		} else {
			baseDistance = jumpers[current-1].minDistance
		}

		// 寻找所有以当前字元开头的分词
		numTokens := 0
		if exclude == "" {
			numTokens = seg.dict.lookupTokens(
				text[current:minInt(current+seg.dict.maxTokenLength, len(text))], tokens, seg.Phrase)
		} else {
			numTokens = seg.dict.lookupTokensExcept(
				text[current:minInt(current+seg.dict.maxTokenLength, len(text))], tokens, seg.Phrase, exclude)
		}
		// 对所有可能的分词，更新分词结束字元处的跳转信息
		//fmt.Printf("new_seg: %s, len=%d\n", textSliceToBytes(text), len(text))
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if !searchMode || current != 0 || location != len(text)-1 {
				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}
		}

		// 当前字元没有对应分词时补加一个伪分词
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: []Text{text[current]}, frequency: 1, distance: 32, pos: "x"})
		}
	}

	// 从后向前扫描第一遍得到需要添加的分词数目
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// 从后向前扫描第二遍添加分词到最终结果
	outputStrings := make([]string, numSeg)
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		if joint == "" {
			outputStrings[numSeg] = jumpers[index].token.Text()
		} else {
			outputStrings[numSeg] = jumpers[index].token.TextOfPhrase(joint)
		}
		index = location - 1
	}
	return outputStrings
}

// 更新跳转信息:
// 	1. 当该位置从未被访问过时(jumper.minDistance为零的情况)，或者
//	2. 当该位置的当前最短路径大于新的最短路径时
// 将当前位置的最短路径值更新为baseDistance加上新分词的概率
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}

// 取两整数较小值
func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// 取两整数较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func notSeparable(prev *rune, cur rune, text Text, pos int) bool {
	if prev == nil {
		return false
	}

	if pos >= len(text) {
		return false
	}

	r, _ := utf8.DecodeRune(text[pos:])
	if ((cur == '.' || cur == '/') && unicode.IsNumber(*prev) && unicode.IsNumber(r)) ||
		(cur == '\'' && unicode.IsLetter(*prev) && unicode.IsLetter(r)) {
		return true
	}

	return false
}

// 将文本划分成字元
func splitTextToWords(text Text, phrase bool) []Text {
	if phrase {
		//if phrase
		output := []Text{}
		rs := []rune(string(text))
		start := 0
		inWord := false
		for i, r := range rs {
			if r == '-' {
				if inWord {
					output = append(output, []byte(string(rs[start:i])))
					inWord = false
				}
				//skip
			} else {
				if !inWord {
					start = i
					inWord = true
				}
				//already in word, do nosplitEnglishTextToWordsthing
			}
		}
		if inWord {
			output = append(output, []byte(string(rs[start:len(rs)])))
		}
		return output
	}
	output := make([]Text, 0, len(text)/3)
	current := 0
	inAlphanumeric := true
	alphanumericStart := 0
	var prev *rune
	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r) ||
			notSeparable(prev, r, text, current+size)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 {
					output = append(output, toLower(text[alphanumericStart:current]))
				}
			}
			output = append(output, text[current:current+size])
		}
		current += size
		prev = &r
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric {
		if current != 0 {
			output = append(output, toLower(text[alphanumericStart:current]))
		}
	}

	return output
}

func splitEnglishTextToWords(text Text, phrase bool) []Text {
	if phrase {
		//if phrase
		output := []Text{}
		rs := []rune(string(text))
		start := 0
		inWord := false
		for i, r := range rs {
			if r == '-' {
				if inWord {
					output = append(output, []byte(string(rs[start:i])))
					inWord = false
				}
				//skip
			} else {
				if !inWord {
					start = i
					inWord = true
				}
				//already in word, do nothing
			}
		}
		if inWord {
			output = append(output, []byte(string(rs[start:len(rs)])))
		}
		return output
	}
	output := make([]Text, 0, len(text)/3)
	current := 0
	inAlphanumeric := true
	alphanumericStart := 0
	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsNumber(r)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 {
					output = append(output, toLower(text[alphanumericStart:current]))
				}
			}
			output = append(output, text[current:current+size])
		}
		current += size
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric {
		if current != 0 {
			output = append(output, toLower(text[alphanumericStart:current]))
		}
	}

	return output
}

// 将英文词转化为小写
func toLower(text []byte) []byte {
	output := make([]byte, len(text))
	for i, t := range text {
		if t >= 'A' && t <= 'Z' {
			output[i] = t - 'A' + 'a'
		} else {
			output[i] = t
		}
	}
	return output
}

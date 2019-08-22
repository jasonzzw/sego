package sego

import (
	"github.com/adamzy/cedar-go"
)

// Dictionary结构体实现了一个字串前缀树，一个分词可能出现在叶子节点也有可能出现在非叶节点
type Dictionary struct {
	trie           *cedar.Cedar // Cedar 前缀树
	maxTokenLength int          // 词典中最长的分词
	tokens         []Token      // 词典中所有的分词，方便遍历
	totalFrequency int64        // 词典中所有分词的频率之和
}

func NewDictionary() *Dictionary {
	return &Dictionary{trie: cedar.New()}
}

// 词典中最长的分词
func (dict *Dictionary) MaxTokenLength() int {
	return dict.maxTokenLength
}

// 词典中分词数目
func (dict *Dictionary) NumTokens() int {
	return len(dict.tokens)
}

// 词典中所有分词的频率之和
func (dict *Dictionary) TotalFrequency() int64 {
	return dict.totalFrequency
}

// 向词典中加入一个分词
func (dict *Dictionary) addToken(token Token, phrase bool) {
	var bytes []byte
	if phrase {
		bytes = textSliceToBytesPhrase(token.text)
	} else {
		bytes = textSliceToBytes(token.text)
	}
	value, err := dict.trie.Get(bytes)
	if err == nil {
		oldFreq := dict.tokens[value].frequency
		if token.frequency > oldFreq {
			dict.tokens[value].frequency = token.frequency
			dict.totalFrequency += int64(token.frequency - oldFreq)
		}
		return
	}

	dict.trie.Insert(bytes, dict.NumTokens())
	dict.tokens = append(dict.tokens, token)
	dict.totalFrequency += int64(token.frequency)
	if len(token.text) > dict.maxTokenLength {
		dict.maxTokenLength = len(token.text)
	}
}

// 在词典中查找和字元组words可以前缀匹配的所有分词
// 返回值为找到的分词数
func (dict *Dictionary) lookupTokens(words []Text, tokens []*Token, phrase bool) (numOfTokens int) {
	var id, value int
	var err error
	for idx, word := range words {
		if idx != 0 && phrase {
			id, err = dict.trie.Jump(append([]byte("-"), word...), id)
		} else {
			id, err = dict.trie.Jump(word, id)
		}

		if err != nil {
			break
		}
		value, err = dict.trie.Value(id)
		if err == nil {
			tokens[numOfTokens] = &dict.tokens[value]
			numOfTokens++
		}
	}
	return
}

func (dict *Dictionary) lookupTokensExcept(words []Text, tokens []*Token, phrase bool, exclude string) (numOfTokens int) {
	var id, value int
	var err error
	for idx, word := range words {
		if idx != 0 && phrase {
			id, err = dict.trie.Jump(append([]byte("-"), word...), id)
		} else {
			id, err = dict.trie.Jump(word, id)
		}

		if err != nil {
			break
		}
		value, err = dict.trie.Value(id)
		if err == nil && dict.tokens[value].Text() != exclude {
			tokens[numOfTokens] = &dict.tokens[value]
			numOfTokens++
		}
	}
	return
}

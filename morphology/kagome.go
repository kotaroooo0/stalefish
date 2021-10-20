package morphology

import (
	ipaneologd "github.com/ikawaha/kagome-dict-ipa-neologd"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// github.com/ikawaha/kagomeに直接依存しないようにラップする
type Kagome struct {
	kagome *tokenizer.Tokenizer
}

func NewKagome() (*Kagome, error) {
	tokenizer, err := tokenizer.New(ipaneologd.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}
	return &Kagome{
		kagome: tokenizer,
	}, nil
}

func (k *Kagome) Analyze(text string) []MorphologyToken {
	tokens := k.kagome.Analyze(text, tokenizer.Search)
	kagomeTokens := make([]MorphologyToken, 0)
	for _, token := range tokens {
		features := token.Features()
		if features[1] == "空白" {
			continue
		}
		kana := token.Surface
		if len(features) >= 8 {
			kana = features[7]
		}
		kagomeTokens = append(kagomeTokens, NewMorphologyToken(token.Surface, kana))
	}
	return kagomeTokens
}

package morphology

type Morphology interface {
	Analyze(string) []MorphologyToken
}

type MorphologyToken struct {
	Term string
	Kana string
}

func NewMorphologyToken(term, kana string) MorphologyToken {
	return MorphologyToken{
		Term: term,
		Kana: kana,
	}
}

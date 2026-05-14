package compress

// Config controls the compression behaviour of [Document].
type Config struct {
	// MinSimilarity is the minimum Jaccard similarity (0..1) for two schemas
	// to be considered for merging. 1.0 means exact equality only (default).
	// Values below 1.0 allow merging schemas with significant structural overlap.
	MinSimilarity float64

	// SimilarityStep is the amount by which the similarity threshold is reduced
	// between rounds when no merges are found at the current threshold.
	// Default is 0.05.
	SimilarityStep float64
}

func (c *Config) setDefaults() {
	if c.MinSimilarity == 0 {
		c.MinSimilarity = 1.0
	}
	if c.SimilarityStep == 0 {
		c.SimilarityStep = 0.05
	}
}

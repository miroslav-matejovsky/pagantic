package rerank

// SelectionPolicy controls filtering after scoring.
type SelectionPolicy struct {
	TopK     int     // max candidates to return, 0 means all
	MinScore float64 // minimum score threshold, 0 means no filter
}

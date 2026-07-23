package models

type DocumentKind string

const (
	DocKindREADME    DocumentKind = "README"
	DocKindDOC       DocumentKind = "DOC"
	DocKindADR       DocumentKind = "ADR"
	DocKindCHANGELOG DocumentKind = "CHANGELOG"
)

//go:generate stringer -type=ExcelPos
package excel_cols

type ExcelPos int

// cols in excel
const (
	A ExcelPos = iota
	B
	C
	D
	E
	F
	G
	H
	I
	J
	K
	L
	M
	N
	O
	P
	Q
	R
	S
	T
	U
	V
	W
	X
	Y
	Z
	AA
	AB
	AC
	AD
	AE
	AF
	AG
	AH
	AI
	AJ
	AK
	AL
	AM
	AN
	AO
	AP
	AQ
	AR
	AS
	AT
	AU
	AV
	AW
	AX
	AY
	AZ
	BA
	BB
	BC
	BD
	BE
	BF
	BG
	BH
	BI
	BJ
	BK
	BL
	BM
	BN
	BO
	BP
	BQ
	BR
	BS
	BT
	BU
	BV
	BW
	BX
	BY
	BZ
	CA
	CB
	CC
	CD
	CE
	CF
	CG
	CH
	CI
	CJ
	CK
	CL
	CM
	CN
	CO
	CP
	CQ
	CR
	CS
	CT
	CU
	CV
	CW
	CX
	CY
	CZ
	DA
	DB
	DC
	DD
	DE
	DF
	DG
)

// maxCols is the number of columns Excel supports, "A" to "XFD"
const maxCols = 16384

// GetPosFromString returns the 0-based position of an Excel column name like
// "A", "B" or "AA". Names beyond the ExcelPos constants above are resolved as
// well, an invalid name returns -1
func GetPosFromString(v string) ExcelPos {
	if v == "" {
		return -1
	}

	pos := 0
	for _, r := range v {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		if r < 'A' || r > 'Z' {
			return -1
		}

		pos = pos*26 + int(r-'A') + 1
		if pos > maxCols {
			return -1
		}
	}

	return ExcelPos(pos - 1)
}

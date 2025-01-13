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

func GetPosFromString(v string) ExcelPos {
	switch v {
	case "A":
		return A
	case "B":
		return B
	case "C":
		return C
	case "D":
		return D
	case "E":
		return E
	case "F":
		return F
	case "G":
		return G
	case "H":
		return H
	case "I":
		return I
	case "J":
		return J
	case "K":
		return K
	case "L":
		return L
	case "M":
		return M
	case "N":
		return N
	case "O":
		return O
	case "P":
		return P
	case "Q":
		return Q
	case "R":
		return R
	case "S":
		return S
	case "T":
		return T
	case "U":
		return U
	case "V":
		return V
	case "W":
		return W
	case "X":
		return X
	case "Y":
		return Y
	case "Z":
		return Z
	case "AA":
		return AA
	case "AB":
		return AB
	case "AC":
		return AC
	case "AD":
		return AD
	case "AE":
		return AE
	case "AF":
		return AF
	case "AG":
		return AG
	case "AH":
		return AH
	case "AI":
		return AI
	case "AJ":
		return AJ
	case "AK":
		return AK
	case "AL":
		return AL
	case "AM":
		return AM
	case "AN":
		return AN
	case "AO":
		return AO
	case "AP":
		return AP
	case "AQ":
		return AQ
	case "AR":
		return AR
	case "AS":
		return AS
	case "AT":
		return AT
	case "AU":
		return AU
	case "AV":
		return AV
	case "AW":
		return AW
	case "AX":
		return AX
	case "AY":
		return AY
	case "AZ":
		return AZ
	case "BA":
		return BA
	case "BB":
		return BB
	case "BC":
		return BC
	case "BD":
		return BD
	case "BE":
		return BE
	case "BF":
		return BF
	case "BG":
		return BG
	case "BH":
		return BH
	case "BI":
		return BI
	case "BJ":
		return BJ
	case "BK":
		return BK
	case "BL":
		return BL
	case "BM":
		return BM
	case "BN":
		return BN
	case "BO":
		return BO
	case "BP":
		return BP
	case "BQ":
		return BQ
	case "BR":
		return BR
	case "BS":
		return BS
	case "BT":
		return BT
	case "BU":
		return BU
	case "BV":
		return BV
	case "BW":
		return BW
	case "BX":
		return BX
	case "BY":
		return BY
	case "BZ":
		return BZ
	case "CA":
		return CA
	case "CB":
		return CB
	case "CC":
		return CC
	case "CD":
		return CD
	case "CE":
		return CE
	case "CF":
		return CF
	case "CG":
		return CG
	case "CH":
		return CH
	case "CI":
		return CI
	case "CV":
		return CV
	case "CW":
		return CW
	case "CX":
		return CX
	case "CY":
		return CY
	case "CZ":
		return CZ
	case "DA":
		return DA
	case "DB":
		return DB
	case "DD":
		return DD
	case "DE":
		return DE
	case "DF":
		return DF
	case "DG":
		return DG
	default:
		return -1
	}
}

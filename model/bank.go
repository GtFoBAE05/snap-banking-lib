package model

type BankCode string

const (
	BankMandiri BankCode = "MANDIRI"
	BankBCA     BankCode = "BCA"
	BankBRI     BankCode = "BRI"
	BankBNI     BankCode = "BNI"
	BankPermata BankCode = "PERMATA"
	BankCIMB    BankCode = "CIMB"
)

func (b BankCode) Name() string {
	switch b {
	case BankMandiri:
		return "Bank Mandiri"
	case BankBCA:
		return "BCA"
	case BankBRI:
		return "BRI"
	case BankBNI:
		return "BNI"
	case BankPermata:
		return "Bank Permata"
	case BankCIMB:
		return "CIMB Niaga"
	default:
		return string(b)
	}
}

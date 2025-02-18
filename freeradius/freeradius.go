package freeradius

import (
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// FreeRADIUS vendor ID.
const (
	VendorID = 11344
)

// Statistics types.
const (
	StatisticsType = 127

	StatisticsTypeNone                  = 0
	StatisticsTypeAuthentication        = 1
	StatisticsTypeAccounting            = 2
	StatisticsTypeProxyAuthentication   = 4
	StatisticsTypeProxyAccounting       = 8
	StatisticsTypeInternal              = 16
	StatisticsTypeClient                = 32
	StatisticsTypeServer                = 64
	StatisticsTypeHomeServer            = 128
	StatisticsTypeAuthAcct              = 3
	StatisticsTypeProxyAuthAcct         = 12
	StatisticsTypeAuthAcctProxyAuthAcct = 15
	StatisticsTypeAll                   = 31
)

// Statistics attributes.
const (
	TotalAccessRequests        = 128
	TotalAccessAccepts         = 129
	TotalAccessRejects         = 130
	TotalAccessChallenges      = 131
	TotalAuthResponses         = 132
	TotalAuthDuplicateRequests = 133
	TotalAuthMalformedRequests = 134
	TotalAuthInvalidRequests   = 135
	TotalAuthDroppedRequests   = 136
	TotalAuthUnknownTypes      = 137

	TotalProxyAccessRequests        = 138
	TotalProxyAccessAccepts         = 139
	TotalProxyAccessRejects         = 140
	TotalProxyAccessChallenges      = 141
	TotalProxyAuthResponses         = 142
	TotalProxyAuthDuplicateRequests = 143
	TotalProxyAuthMalformedRequests = 144
	TotalProxyAuthInvalidRequests   = 145
	TotalProxyAuthDroppedRequests   = 146
	TotalProxyAuthUnknownTypes      = 147

	TotalAccountingRequests    = 148
	TotalAccountingResponses   = 149
	TotalAcctDuplicateRequests = 150
	TotalAcctMalformedRequests = 151
	TotalAcctInvalidRequests   = 152
	TotalAcctDroppedRequests   = 153
	TotalAcctUnknownTypes      = 154

	TotalProxyAccountingRequests    = 155
	TotalProxyAccountingResponses   = 156
	TotalProxyAcctDuplicateRequests = 157
	TotalProxyAcctMalformedRequests = 158
	TotalProxyAcctInvalidRequests   = 159
	TotalProxyAcctDroppedRequests   = 160
	TotalProxyAcctUnknownTypes      = 161
)

// GetValue returns attribute value.
func GetValue(p *radius.Packet, typ byte) (value uint32, err error) {
	a, ok := lookupVendor(p, typ)
	if !ok {
		err = radius.ErrNoAttribute
		return
	}
	value, err = radius.Integer(a)
	return
}

// SetValue sets attribute value.
func SetValue(p *radius.Packet, typ byte, attr radius.Attribute) (err error) {
	err = setVendor(p, typ, attr)
	return
}

func addVendor(p *radius.Packet, typ byte, attr radius.Attribute) (err error) {
	var vsa radius.Attribute
	vendor := make(radius.Attribute, 2+len(attr))
	vendor[0] = typ
	vendor[1] = byte(len(vendor))
	copy(vendor[2:], attr)
	vsa, err = radius.NewVendorSpecific(VendorID, vendor)
	if err != nil {
		return
	}
	p.Add(rfc2865.VendorSpecific_Type, vsa)
	return
}

func setVendor(p *radius.Packet, typ byte, attr radius.Attribute) (err error) {
	for i := 0; i < len(p.Attributes[rfc2865.VendorSpecific_Type]); {
		vendorID, vsa, err := radius.VendorSpecific(p.Attributes[rfc2865.VendorSpecific_Type][i])
		if err != nil || vendorID != VendorID {
			i++
			continue
		}
		for j := 0; len(vsa[j:]) >= 3; {
			vsaTyp, vsaLen := vsa[0], vsa[1]
			if int(vsaLen) > len(vsa[j:]) || vsaLen < 3 {
				i++
				break
			}
			if vsaTyp == typ {
				vsa = append(vsa[:j], vsa[j+int(vsaLen):]...)
			}
			j += int(vsaLen)
		}
		if len(vsa) > 0 {
			copy(p.Attributes[rfc2865.VendorSpecific_Type][i][4:], vsa)
			i++
		} else {
			p.Attributes[rfc2865.VendorSpecific_Type] = append(p.Attributes[rfc2865.VendorSpecific_Type][:i], p.Attributes[rfc2865.VendorSpecific_Type][i+i:]...)
		}
	}
	return addVendor(p, typ, attr)
}

func lookupVendor(p *radius.Packet, typ byte) (attr radius.Attribute, ok bool) {
	for _, a := range p.Attributes[rfc2865.VendorSpecific_Type] {
		vendorID, vsa, err := radius.VendorSpecific(a)
		if err != nil || vendorID != VendorID {
			continue
		}
		for len(vsa) >= 3 {
			vsaTyp, vsaLen := vsa[0], vsa[1]
			if int(vsaLen) > len(vsa) || vsaLen < 3 {
				break
			}
			if vsaTyp == typ {
				return vsa[2:int(vsaLen)], true
			}
			vsa = vsa[int(vsaLen):]
		}
	}
	return
}

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package roomupdatefbs

import "strconv"

type UpdateType int8

const (
	UpdateTypePing              UpdateType = 0
	UpdateTypeClaim             UpdateType = 1
	UpdateTypeRetract           UpdateType = 2
	UpdateTypeSubscribe         UpdateType = 3
	UpdateTypeDeath             UpdateType = 4
	UpdateTypeSubscriptionDeath UpdateType = 5
)

var EnumNamesUpdateType = map[UpdateType]string{
	UpdateTypePing:              "Ping",
	UpdateTypeClaim:             "Claim",
	UpdateTypeRetract:           "Retract",
	UpdateTypeSubscribe:         "Subscribe",
	UpdateTypeDeath:             "Death",
	UpdateTypeSubscriptionDeath: "SubscriptionDeath",
}

var EnumValuesUpdateType = map[string]UpdateType{
	"Ping":              UpdateTypePing,
	"Claim":             UpdateTypeClaim,
	"Retract":           UpdateTypeRetract,
	"Subscribe":         UpdateTypeSubscribe,
	"Death":             UpdateTypeDeath,
	"SubscriptionDeath": UpdateTypeSubscriptionDeath,
}

func (v UpdateType) String() string {
	if s, ok := EnumNamesUpdateType[v]; ok {
		return s
	}
	return "UpdateType(" + strconv.FormatInt(int64(v), 10) + ")"
}

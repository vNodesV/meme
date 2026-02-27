package types

import proto "github.com/cosmos/gogoproto/proto"

func init() {
	// Register human-readable aliases for AccessType in the global proto enum
	// registry so gogoproto/jsonpb can unmarshal legacy JSON that uses the
	// human-readable names ("Everybody", "Nobody", "OnlyAddress") instead of
	// the canonical proto names ("ACCESS_TYPE_EVERYBODY" etc.).
	// NOTE: This init() must run AFTER params.pb.go registers the enum.
	// go build processes files alphabetically: params.pb.go < params_enum_aliases.go,
	// but within the same package, init ordering matches file order so this runs after.
	if m := proto.EnumValueMap("cosmwasm.wasm.v1.AccessType"); m != nil {
		m["Unspecified"] = int32(AccessTypeUnspecified)
		m["Nobody"] = int32(AccessTypeNobody)
		m["OnlyAddress"] = int32(AccessTypeOnlyAddress)
		m["Everybody"] = int32(AccessTypeEverybody)
	}
}

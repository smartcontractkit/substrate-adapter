package adapter

import (
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

// removeHexPrefix removes the prefix (0x) of a given hex string.
func removeHexPrefix(str string) string {
	if hasHexPrefix(str) {
		return str[2:]
	}
	return str
}

// hasHexPrefix returns true if the string starts with 0x.
func hasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func parseDecimalString(input string) (*big.Int, error) {
	parseValue, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return nil, err
	}
	output, ok := big.NewInt(0).SetString(fmt.Sprintf("%.f", parseValue), 10)
	if !ok {
		return nil, fmt.Errorf("error parsing decimal %s", input)
	}
	return output, nil
}

func ParseNumericString(input string) (decimal.Decimal, error) {
	if hasHexPrefix(input) {
		output, ok := big.NewInt(0).SetString(removeHexPrefix(input), 16)
		if !ok {
			return decimal.Decimal{}, fmt.Errorf("error parsing hex %s", input)
		}
		return decimal.NewFromBigInt(output, 0), nil
	}

	return decimal.NewFromString(input)
}

func convertTypes(t, v string) (interface{}, error) {
	switch strings.ToLower(t) {
	case "bool":
		switch strings.ToLower(v) {
		case "true":
			return types.NewBool(true), nil
		case "false":
			return types.NewBool(false), nil
		default:
			return nil, errors.New("unable to parse bool")
		}
	case "uint8":
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing uint8")
		}
		return types.NewU8(uint8(i)), nil
	case "uint16":
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing uint16")
		}
		return types.NewU16(uint16(i)), nil
	case "uint32":
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing uint32")
		}
		return types.NewU32(uint32(i)), nil
	case "uint64":
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing uint64")
		}
		return types.NewU64(i), nil
	case "int8":
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing int8")
		}
		return types.NewI8(int8(i)), nil
	case "int16":
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing int16")
		}
		return types.NewI16(int16(i)), nil
	case "int32":
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing int32")
		}
		return types.NewI32(int32(i)), nil
	case "int64":
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing int64")
		}
		return types.NewI64(i), nil
	case "int128", "uint128", "int256", "uint256":
		i, err := ParseNumericString(v)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing numeric string")
		}
		switch strings.ToLower(t) {
		case "int128":
			return types.NewI128(*i.BigInt()), nil
		case "uint128":
			return types.NewU128(*i.BigInt()), nil
		case "int256":
			return types.NewI256(*i.BigInt()), nil
		case "uint256":
			return types.NewU256(*i.BigInt()), nil
		}
	case "ucompact":
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed parsing ucompact")
		}
		return types.NewUCompact(new(big.Int).SetInt64(i)), nil
	case "bytes":
		return types.Bytes(v), nil
	case "address":
		return types.NewAddressFromHexAccountID(fmt.Sprintf("%s", v))
	}

	return nil, errors.New("unknown type")
}

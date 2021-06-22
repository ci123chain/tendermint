package vrf

import "errors"

var (
	// ErrPointNotOnCurve occurs when a public key is not on the curve.
	ErrPointNotOnCurve = errors.New("point is not on the P256 curve")
	// ErrWrongKeyType occurs when a key is not an ECDSA key.
	ErrWrongKeyType = errors.New("not an ECDSA key")
	// ErrNoPEMFound occurs when attempting to parse a non PEM data structure.
	ErrNoPEMFound = errors.New("no PEM block found")
	// ErrInvalidVRF occurs when the VRF does not validate.
	ErrInvalidVRF = errors.New("invalid VRF proof")
	// ErrEvaluateFailed fail
	ErrEvaluateFailed = errors.New("failed to evaluate VRF")
)

package supervisor

//-----------------------------------------------------------------------------

// errString a string that inplements the error interface
type errString string

func (v errString) Error() string { return string(v) }

//-----------------------------------------------------------------------------

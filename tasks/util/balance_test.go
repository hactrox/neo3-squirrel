package util

import "testing"

func TestGenerateNEP17BalanceOfScript(t *testing.T) {
	addr := "NgPkjjLTNcQad99iRYeXRUuowE4gxLAnDL"
	contract := "0x668e0c1f9d7b70a99dd9e06eadd4c784d641afbc"
	want := "0c14e0a3c55cad72028fb5901748b19a27be21f6540411c01f0c0962616c616e63654f660c14bcaf41d684c7d4ad6ee0d99da9707b9d1f0c8e6641627d5b52"

	get, err := generateNEP17BalanceOfScript(addr, contract)
	if err != nil {
		t.Fatal(err)
	}

	if get != want {
		t.Fatalf("Incorrect result from generateNEP17BalanceOfScript\n get: %s\nwant: %s", get, want)
	}
}

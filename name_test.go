package main

import (
	"testing"
)

func TestParseProtoName(t *testing.T) {
	var testData = []struct{ fullname, expected string }{
		{"__ZN3abc3def4testEv", "abcdeftest"},       // abc::def::test
		{"__ZN3abc3def5_testEv", "abcdef_test"},     // abc::def::_test
		{"__ZN3abc3def6__testEv", "abcdef__test"},   // abc::def::__test
		{"__ZN3abc3def7___testEv", "abcdef___test"}, // abc::def::___test
		{"__ZN3abc4testEv", "abctest"},              // abc::test
		{"__ZN3abc5_testEv", "abc_test"},            // abc::_test
		{"__ZN3abc6__testEv", "abc__test"},          // abc::__test
		{"__ZN3abc7___testEv", "abc___test"},        // abc::___test
		{"_ZN3abc3def4testEv", "abcdeftest"},        // abc::def::test
		{"_ZN3abc3def5_testEv", "abcdef_test"},      // abc::def::_test
		{"_ZN3abc3def6__testEv", "abcdef__test"},    // abc::def::__test
		{"_ZN3abc3def7___testEv", "abcdef___test"},  // abc::def::___test
		{"_ZN3abc4testEv", "abctest"},               // abc::test
		{"_ZN3abc5_testEv", "abc_test"},             // abc::_test
		{"_ZN3abc6__testEv", "abc__test"},           // abc::__test
		{"_ZN3abc7___testEv", "abc___test"},         // abc::___test

		{"XORShift128Plus", "XORShift128Plus"},
		{"test", "test"},
		{"_test", "test"},
		{"__test", "_test"}, // don's use this

		{"_ZN4Simd4Avx213Yuv444pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv444pToBgra"}, // Simd::Avx2::Yuv444pToBgra
		{"_ZN4Simd4Avx213Yuv420pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv420pToBgra"}, // Simd::Avx2::Yuv420pToBgra
		{"_ZN4Simd4Avx213Yuv422pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv422pToBgra"}, // Simd::Avx2::Yuv422pToBgra
		{"_ZN4Simd4Avx213ReduceGray2x2EPKhmmmPhmmm", "SimdAvx2ReduceGray2x2"},        // Simd::Avx2::ReduceGray2x2
		{"_ZN4Simd4Avx216AbsDifferenceSumEPKhmS2_mmmPy", "SimdAvx2AbsDifferenceSum"}, // Simd::Avx2::AbsDifferenceSum

		{"__ZN4Simd4Avx213Yuv444pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv444pToBgra"}, // Simd::Avx2::Yuv444pToBgra
		{"__ZN4Simd4Avx213Yuv420pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv420pToBgra"}, // Simd::Avx2::Yuv420pToBgra
		{"__ZN4Simd4Avx213Yuv422pToBgraEPKhmS2_mS2_mmmPhmh", "SimdAvx2Yuv422pToBgra"}, // Simd::Avx2::Yuv422pToBgra
		{"__ZN4Simd4Avx213ReduceGray2x2EPKhmmmPhmmm", "SimdAvx2ReduceGray2x2"},        // Simd::Avx2::ReduceGray2x2
		{"__ZN4Simd4Avx216AbsDifferenceSumEPKhmS2_mmmPy", "SimdAvx2AbsDifferenceSum"}, // Simd::Avx2::AbsDifferenceSum
	}

	for i := range testData {
		name := parseProtoName(testData[i].fullname)
		if name != testData[i].expected {
			t.Errorf("%02d: expected='%s', got='%s'", i, testData[i].expected, name)
		}
	}
}

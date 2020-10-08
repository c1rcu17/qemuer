package util

import (
	"fmt"
	"math/big"
	"net"
)

func AddressRange(subnet *net.IPNet) (gw net.IP, bc net.IP, start net.IP, end net.IP, err error) {
	prefixLen, bits := subnet.Mask.Size()
	hostLen := uint(bits) - uint(prefixLen)

	subnetInt := (&big.Int{}).SetBytes([]byte(subnet.IP))

	gwInt := (&big.Int{}).Set(subnetInt)
	gwInt.Add(gwInt, big.NewInt(1))

	bcInt := big.NewInt(1)
	bcInt.Lsh(bcInt, hostLen)
	bcInt.Sub(bcInt, big.NewInt(1))
	bcInt.Or(bcInt, subnetInt)

	startInt := (&big.Int{}).Set(gwInt)
	startInt.Add(startInt, big.NewInt(1))

	endInt := (&big.Int{}).Set(bcInt)
	endInt.Sub(endInt, big.NewInt(1))

	if startInt.Cmp(bcInt) > 0 {
		err = fmt.Errorf("invalid netmask %s: too narrow", net.IP(subnet.Mask).String())
		return
	}

	gw = intToIP(gwInt, bits)
	bc = intToIP(bcInt, bits)
	start = intToIP(startInt, bits)
	end = intToIP(endInt, bits)

	return
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)

	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}

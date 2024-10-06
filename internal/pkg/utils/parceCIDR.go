package utils

import "net"

// GetCIDR парсит строку CIDR для получения подсети
func GetCIDR(trustedString string) (*net.IPNet, error) {
	if trustedString == "" {
		return nil, nil
	}
	//нам интересна подсеть
	_, ipNet, err := net.ParseCIDR(trustedString)
	if err != nil {
		return nil, err
	}
	return ipNet, nil
}

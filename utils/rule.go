package utils

func FofaRules() string {
	return `&& server!="cloudflare" && server!="alicdn" && server!="qcloud"` +
		` && server!="yunjiasu" && server!="yupaicloud" && cloud_name!="Cloudflare"` +
		` && server!="upyun" && server!="ws" && server!="cdnws"` +
		` && server!="china cache" && server!="fastly" && server!="akamai"` +
		` && server!="akamaighost" && server!="cloudfront" && server!="hwcdn"` +
		` && server!="wangzhansheshi" && server!="360wzws" && server!="incapsula"` +
		` && server!="stackpath" && server!="keycdn" && cloud_name!="cloudfront"` +
		` && org!="CLOUDFLARENET" && server!="layun.com" && server!="*cdn*" && server!="Byte-nginx"`
}

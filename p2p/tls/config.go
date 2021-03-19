package tls

type TLSConfig struct {
	RemoteTLSCertURI            string `mapstructure:"remote_tls_cert"`
	RemoteTLSCertKeyURI         string `mapstructure:"remote_tls_cert_key"`
	RemoteTLSDialTimeout        int    `mapstructure:"remote_tls_dial_timeout"`
	RemoteTLSInsecureSkipVerify bool   `mapstructure:"remote_tls_insecure_skip_verify"` ///true means close verify.
}

func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		RemoteTLSCertURI:            "",
		RemoteTLSCertKeyURI:         "",
		RemoteTLSDialTimeout:        5,
		RemoteTLSInsecureSkipVerify: true,
	}
}

package core

type XProviderSetupParams struct {
	Provider string
	Region   string
	Zone     string
	App      string
	IpGroup  string
	IpSubnet string
}

const xProviderSetupParam = `
apiVersion: xrds.skycluster.savitestbed.ca/v1alpha1
metadata:
  name: xprovidersetup1-{{.Provider}}-{{.Region}}-{{.Zone}}-{{.App}}
  annotations:
    crossplane.io/paused: "true"
  labels:
    managed-by: skycluster
kind: XProviderSetup
spec:
  forProvider:
    vpnServer:
      host: https://vpn.skycluster.savitestbed.ca
      port: 8080
      token: e37cddbe197d0ce928997e91255ed27564af46d423a3f0de
    gateway:
      flavor: small
      image: ubuntu-22.04
    publicKey: |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDEGdP3tmn2XZ43QqkB92fp03WskHXS1hAnmqOuoYoKtn1LWSXcjbF6oMI/yQErWWi07DfqZm6ziQwKasOn8aVedkVLf0vIEiGGZZMzjh8sv/t+zcWmtFyW4Dcm2qiYXk5ckdzxoPXUpzsx6PwmGpOnV2YKBoX5p1ItyNN1+ltMbv5FCelJR3AWCIqq5LtfiHWZcj/77xyjIFsYA2ZREPN9UySZvJHdsMOHFXTZJq615qL2poG09sFdM2HSrKR7WX/duqm732gpScu0svPpwztQSQY01O4iyx/0X21v4FV5E3/NSM8EMVRfE4i7WfPEajN6PZHPKS/qejTMpgsKswJIO4FLUlDzOKUqvMPW+/sJ3VX5bejAbNdvvu0xz0qBZ5etzCjFxIE2pCP+GaSjfMef1RRd2Q1NEiPPIx3WDBdRN3aKmhfAYfQypJIMjDTMVW1slhhSB6MPibPxXSUm2HnAA+HfrwJXJ9dFLaBcGOyZdAMkYwwCh4dRSg8jnBz3Gic= ubuntu@esn-skycluster-1
    # if this is savi, make sure scinet is in the range 10.30.*.*
    # and vaughan in the range 10.32.*.*
    ipCidrRange: 10.{{.IpGroup}}.{{.IpSubnet}}.0/24
    secgroup:
      description: "SkyCluster VPN Server"
      tcpPorts: 
        - fromPort: 22
          toPort: 22
        - fromPort: 80
          toPort: 80
        - fromPort: 443
          toPort: 443
        - fromPort: 8080
          toPort: 8080
        - fromPort: 6443
          toPort: 6443
      udpPorts: 
        - fromPort: 3478 # stun
          toPort: 3478
        - fromPort: 41641 
          toPort: 41641
  provider:
    name: {{.Provider}}
    region: {{.Region}}
    zone: "{{.Zone}}"
`

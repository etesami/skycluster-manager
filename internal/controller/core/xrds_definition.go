package core

type skyProviderSetupParams struct {
	Name       string
	APIVersion string
	Provider   string
	Region     string
	Zone       string
	AppName    string
	IpGroup    string
	IpSubnet   string
}

type skyK8SClusterSetupParams struct {
	Name        string
	AppName     string
	CtrlNode    skyK8SNode
	WorkerNodes []skyK8SNode
}

type skyK8SNode struct {
	Size     string
	Provider string
	Region   string
}

const skyK8SClusterSetupTemplate = `
apiVersion: xrds.skycluster.savitestbed.ca/v1alpha1
kind: SkyK8SCluster
metadata:
  name: skyk8scluster-{{.AppName}}
  labels:
    managed-by: skycluster
    skycluster/app-name: {{.AppName}}
    skycluster/environment: dev
spec: 
  forProvider:
    ctrl:
      flavor: {{.CtrlNode.Size}}
      image: ubuntu-22.04
      provider:
        name: {{.CtrlNode.Provider}}
        region: {{.CtrlNode.Region}}
        zone: default
    agents: 
    {{- range .WorkerNodes }}
      - flavor: {{.Size}}
        image: ubuntu-22.04
        provider:
          name: {{.Provider}}
          region: {{.Region}}
          zone: default
    {{- end }}
`

const skyProviderSetupTemplate = `
apiVersion: xrds.skycluster.savitestbed.ca/v1alpha1
metadata:
  name: {{.Name}}
  labels:
    managed-by: skycluster
    skycluster/environment: dev
    skycluster/app-name: {{.AppName}}
    skycluster/provider-name: {{.Provider}}
    skycluster/provider-region: {{.Region}}
    skycluster/provider-zone: {{.Zone}}
kind: SkyProviderSetup
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

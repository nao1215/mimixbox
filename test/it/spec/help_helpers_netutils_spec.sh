# Per-command --help contract specs backed by dedicated netutils helpers (issue #489).
Describe 'netutils commands expose a dedicated --help helper'
    Include netutils/arp_test.sh
    Include netutils/arping_test.sh
    Include netutils/brctl_test.sh
    Include netutils/dhcprelay_test.sh
    Include netutils/dnsd_test.sh
    Include netutils/dnsdomainname_test.sh
    Include netutils/dumpleases_test.sh
    Include netutils/ether-wake_test.sh
    Include netutils/fakeidentd_test.sh
    Include netutils/ftpd_test.sh
    Include netutils/ftpget_test.sh
    Include netutils/ftpput_test.sh
    Include netutils/http-status-code_test.sh
    Include netutils/httpd_test.sh
    Include netutils/ifconfig_test.sh
    Include netutils/ifdown_test.sh
    Include netutils/ifenslave_test.sh
    Include netutils/ifplugd_test.sh
    Include netutils/ifup_test.sh
    Include netutils/inetd_test.sh
    Include netutils/ip_test.sh
    Include netutils/ipaddr_test.sh
    Include netutils/iplink_test.sh
    Include netutils/ipneigh_test.sh
    Include netutils/iproute_test.sh
    Include netutils/iprule_test.sh
    Include netutils/iptunnel_test.sh
    Include netutils/nameif_test.sh
    Include netutils/nbd-client_test.sh
    Include netutils/netcat_test.sh
    Include netutils/netstat_test.sh
    Include netutils/nslookup_test.sh
    Include netutils/ntpd_test.sh
    Include netutils/ping6_test.sh
    Include netutils/pscan_test.sh
    Include netutils/route_test.sh
    Include netutils/slattach_test.sh
    Include netutils/ssl_client_test.sh
    Include netutils/ssl_server_test.sh
    Include netutils/tc_test.sh
    Include netutils/tcpsvd_test.sh
    Include netutils/telnet_test.sh
    Include netutils/telnetd_test.sh
    Include netutils/tftp_test.sh
    Include netutils/tftpd_test.sh
    Include netutils/traceroute_test.sh
    Include netutils/traceroute6_test.sh
    Include netutils/tunctl_test.sh
    Include netutils/udhcpc_test.sh
    Include netutils/udhcpc6_test.sh
    Include netutils/udhcpd_test.sh
    Include netutils/udpsvd_test.sh
    Include netutils/vconfig_test.sh
    Include netutils/whois_test.sh
    Include netutils/zcip_test.sh

    It 'arp --help is structured'
        When call ArpHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'arping --help is structured'
        When call ArpingHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'brctl --help is structured'
        When call BrctlHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dhcprelay --help is structured'
        When call DhcprelayHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dnsd --help is structured'
        When call DnsdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dnsdomainname --help is structured'
        When call DnsdomainnameHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'dumpleases --help is structured'
        When call DumpleasesHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ether-wake --help is structured'
        When call EtherWakeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'fakeidentd --help is structured'
        When call FakeidentdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ftpd --help is structured'
        When call FtpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ftpget --help is structured'
        When call FtpgetHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ftpput --help is structured'
        When call FtpputHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'http-status-code --help is structured'
        When call HttpStatusCodeHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'httpd --help is structured'
        When call HttpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ifconfig --help is structured'
        When call IfconfigHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ifdown --help is structured'
        When call IfdownHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ifenslave --help is structured'
        When call IfenslaveHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ifplugd --help is structured'
        When call IfplugdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ifup --help is structured'
        When call IfupHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'inetd --help is structured'
        When call InetdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ip --help is structured'
        When call IpHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ipaddr --help is structured'
        When call IpaddrHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'iplink --help is structured'
        When call IplinkHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ipneigh --help is structured'
        When call IpneighHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'iproute --help is structured'
        When call IprouteHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'iprule --help is structured'
        When call IpruleHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'iptunnel --help is structured'
        When call IptunnelHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'nameif --help is structured'
        When call NameifHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'nbd-client --help is structured'
        When call NbdClientHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'netcat --help is structured'
        When call NetcatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'netstat --help is structured'
        When call NetstatHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'nslookup --help is structured'
        When call NslookupHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ntpd --help is structured'
        When call NtpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ping6 --help is structured'
        When call Ping6Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'pscan --help is structured'
        When call PscanHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'route --help is structured'
        When call RouteHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'slattach --help is structured'
        When call SlattachHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ssl_client --help is structured'
        When call SslClientHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'ssl_server --help is structured'
        When call SslServerHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'tc --help is structured'
        When call TcHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'tcpsvd --help is structured'
        When call TcpsvdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'telnet --help is structured'
        When call TelnetHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'telnetd --help is structured'
        When call TelnetdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'tftp --help is structured'
        When call TftpHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'tftpd --help is structured'
        When call TftpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'traceroute --help is structured'
        When call TracerouteHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'traceroute6 --help is structured'
        When call Traceroute6Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'tunctl --help is structured'
        When call TunctlHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'udhcpc --help is structured'
        When call UdhcpcHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'udhcpc6 --help is structured'
        When call Udhcpc6Help
        The status should be success
        The output should include 'Usage:'
    End
    It 'udhcpd --help is structured'
        When call UdhcpdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'udpsvd --help is structured'
        When call UdpsvdHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'vconfig --help is structured'
        When call VconfigHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'whois --help is structured'
        When call WhoisHelp
        The status should be success
        The output should include 'Usage:'
    End
    It 'zcip --help is structured'
        When call ZcipHelp
        The status should be success
        The output should include 'Usage:'
    End
End

#include <stdio.h>
#include <stdlib.h>

int main(int argc, char **argv) {
    char *main_interface=argv[1];
    char *vm_interface=argv[2];
    char *CIDR=argv[3];

    char comm[1000];
    snprintf(comm, sizeof(comm), "ip tuntap add %s mode tap",vm_interface);
    system(comm);
    snprintf(comm, sizeof(comm), "ip addr add %s dev %s",CIDR,vm_interface);
    system(comm);
    snprintf(comm, sizeof(comm), "ip link set %s up",vm_interface);
    system(comm);
    snprintf(comm, sizeof(comm), "echo 1 > /proc/sys/net/ipv4/ip_forward");
    system(comm);
    snprintf(comm, sizeof(comm), "iptables -t nat -C POSTROUTING -o %s -j MASQUERADE || iptables -t nat -A POSTROUTING -o %s -j MASQUERADE ",main_interface,main_interface);
    system(comm);
    snprintf(comm, sizeof(comm), "iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT || iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT ");
    system(comm);
    snprintf(comm, sizeof(comm), "iptables -C FORWARD -i %s -o %s -j ACCEPT || iptables -A FORWARD -i %s -o %s -j ACCEPT",vm_interface,main_interface,vm_interface,main_interface);
    system(comm);
    return 0;
}

// to run: ./program eth0 tap0 172.16.0.1/24

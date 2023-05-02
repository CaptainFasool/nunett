# What is the Device Management Service (DMS)?

A **device management service** or **DMS** is a program that helps users run various computational services, including machine learning (ML) jobs, on a compute provider machine, based on an NTX token request system on NuNet. In simple terms, it connects users who want to perform computational tasks to powerful CPU/GPU enabled computers that can handle these tasks. The purpose of the DMS is to connect users on NuNet, allow them to run any service (not only ML jobs) and be rewarded for it.

The NTX token is a digital cryptographic asset available on the Cardano and Ethereum blockchain as a smart contract. However, for the current use case of running machine learning jobs, only the Cardano blockchain is being used. Users request and allocate resources for computational jobs through a Service Provider Dashboard. Compute providers receive the NTX tokens based on the jobs through a Compute Provider Dashboard.

Please note that the dashboards are not components of NuNet's core architecture. Both these components have been developed to perform the current use case that is to run ML jobs on compute providers machines.

Here's a step-by-step explanation:

1. Users have computational services they want to run. These services often require a lot of computing power, which may not be available on their personal devices.
2. Compute provider machines are powerful computers designed to handle resource-intensive tasks like machine learning jobs.
3. The device management service acts as a bridge, connecting users with these compute provider machines.
4. Users specify resources and job requirements through a webapp interface, and request access to the compute provider machines by sending NTX tokens. NTX acts as a digital ticket, granting users access to the resources they need.
5. The device management service receives the job request after verifying the authenticity of the NTX transaction through an Oracle.
6. Once received, the DMS allocates the necessary resources on the compute provider machine to run the user's job.
7. The user's job is executed on the provider's machine, and the results are sent back to the user.

In summary, a device management service simplifies the process of running various computational services on powerful computers. Users can easily request access to these resources with NTX tokens, allowing them to complete their tasks efficiently and effectively.

**Note**: If you are a developer, please check out [these instructions](https://gitlab.com/nunet/device-management-service/-/blob/develop/README-DEV.md).

# How to Install the Device Management Service?

Before going through the installation process, let's take a quick look at the system requirements and other things to keep in mind.

## Installing via Virtual Machines or Windows Subsystem (WSL) for Linux

When using a VM or WSL, using Ubuntu 20.04 is highly recommended.

### Things to keep in mind for VMs

- Skip doing an [unattended installation](https://www.virtualbox.org/manual/ch01.html#create-vm-wizard-unattended-install) for the new Ubuntu VM as it might not add the user with administrative privileges.
- Enable [Guest Additions](https://www.virtualbox.org/manual/ch04.html) when installing the VM (VirtualBox only).
- Always [change the default NAT network setting to Bridged](https://www.techrepublic.com/article/how-to-set-bridged-networking-in-a-virtualbox-virtual-machine) before booting the VM.
- [Install Extension Pack](https://phoenixnap.com/kb/install-virtualbox-extension-pack) if on VirtualBox (recommended)
- [Install VMware Tools](https://kb.vmware.com/s/article/1014294) if on VMware (recommended)
- ML on GPU jobs on VMs are not supported

### Things to keep in mind for WSLs

- Install WSL through the Windows Store.
- Install the [Update KB5020030](https://www.catalog.update.microsoft.com/Search.aspx?q=KB5020030) (Windows 10 only)
- Install Ubuntu 20.04 through WSL
- Enable [systemd on Ubuntu WSL](https://www.xda-developers.com/how-enable-systemd-in-wsl)
- ML Jobs deployed on Linux cannot be resumed on WSL

Though it is possible to run ML jobs on Windows machines with WSL, using Ubuntu 20.04 natively is highly recommended. The reason being our development is completely based around the Linux operating system. Also, the system requirements when using WSL would increase by at least around 25%.

If you are using a dual boot machine, make sure you use the `wsl --shutdown` command before shutting down Windows and running Linux for ML jobs. Also, ensure your Windows machine is not in a hibernated state when you reboot into Linux.

## CPU-only machines

### Minimum System requirements

We only require for you to specify CPU (MHz x no. of cores) and RAM but your system must meet at least the following set of requirements before you decide to onboard it:

- CPU - 2 GHz
- RAM - 4 GB
- Free Disk Space - 10 GB
- Internet Download/Upload Speed - 4 Mbps / 0.5 MBps

If the above CPU has 4 cores, your available CPU would be around 8000 MHz. So if you want to onboard half your CPU and RAM on NuNet, you can specify 4000 MHz CPU and 2000 MB RAM.

### Recommended System requirements

- CPU - 3.5 GHz
- RAM - 8-16 GB
- Free Disk Space - 20 GB
- Internet Download/Upload Speed - 10 Mbps / 1.25 MBps

## GPU Machines

### Minimum System Requirements

- CPU - 3 GHz
- RAM - 8 GB
- NVIDIA GPU - 4 GB VRAM
- Free Disk Space - 50 GB
- Internet Download/Upload Speed - 50 Mbps

### Recommended System requirements

- CPU - 4 GHz
- RAM - 16-32 GB
- NVIDIA GPU - 8-12 GB VRAM
- Free Disk Space - 100 GB
- Internet Download/Upload Speed - 100 Mbps

Here's a step by step process to install the device management service (DMS) on a compute provider machine:

1. **Download the DMS package**:

   Download the latest version with this command:

```
wget https://d.nunet.io/nunet-dms-latest.deb -O nunet-dms-latest.deb
```

2. **Install DMS**: 

   DMS has some dependencies, but they'll be installed automatically during the installation process.

​	   Open a terminal and navigate to the directory where you downloaded the DMS package (skip this step if you used 	   the ***wget*** command above). Install the DMS with this command:

```
sudo apt update && sudo apt install ./nunet-dms-latest.deb -y
```

​		If the installation fails, try these commands instead:

```
sudo dpkg -i nunet-dms-latest.deb
sudo apt -f install -y
```

If you see a "Permission denied" error, don't worry, it's just a notice. Proceed to the next step.

Check if DMS is running: Look for "/usr/bin/nunet-dms" in the output of this command:

```
ps aux | grep nunet-dms
```

If it's not running, [submit a bug report](https://gitlab.com/nunet/documentation/-/issues) with the error messages. Here are the [contribution guidelines](https://gitlab.com/nunet/documentation/-/wikis/Contribution-Guidelines).

3. **Uninstall DMS** (if needed): 

To remove DMS, use this command:

```
sudo apt remove nunet-dms
```

To download and install a new DMS package, repeat steps 1 and 2.

4. **Completely remove DMS** (if needed): 

To fully uninstall and stop DMS, use either of these commands:

```

sudo apt purge nunet-dms
```

or

```
sudo dpkg --purge nunet-dms
```

5. **Update DMS**: 

To update the DMS to the latest version, follow these steps in the given sequence:

​	a. Uninstall the current DMS (Step 3) 
​	b. Download the latest DMS package (Step 1) 
​	c. Install the new DMS package (Step 2)


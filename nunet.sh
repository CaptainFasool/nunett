echo "+++++++++++++++++++++++++++++++++++++++++++++++"
echo "\033[0;35m"
echo "    _____                       _   _        ";
echo "   / ____|                     | | (_)       ";
echo "  | (___   ___ _ __ _   _  ___ | |_ _  ___   ";
echo "   \___ \ / __| '__| | | |/ _ \| __| |/ __|  ";
echo "   ____) | (__| |  | |_| | (_) | |_| | (__   ";
echo "  |_____/ \___|_|   \__, |\___/ \__|_|\___|  ";
echo "                     __/ |                   ";
echo "                    |___/                    ";
echo "\033[0m"
echo "+++++++++++++++++++++++++++++++++++++++++++++++"

sleep 2

echo -e "\e[1m\e[32m1. Downloading DMS... \e[0m" && sleep 1
cd $HOME && wget https://d.nunet.io/nunet-dms-latest.deb

echo -e "\e[1m\e[32m2. Installing DMS... \e[0m" && sleep 1
sudo apt update && sudo apt install ./nunet-dms-latest.deb

echo -e "\e[1m\e[32m3. DMS status... \e[0m" && sleep 1
systemctl show -p SubState nunet-dms | sed 's/SubState=//g'

sleep 2

while true; do
	read -p "Do you have 0x Address? (yes/no)" nWalResp
	case $nWalResp in
		yes ) break ;;
		no  ) 
    		echo -e "\e[1m\e[32mCreating New Wallet...  \e[0m" && sleep 1
			nWalletAddress=$(nunet wallet new)
			nAddress=$(echo $nWalletAddress | jq -r .address)
			echo "Your new wallet information used for NuNet: "
			echo -e "Address: \033[33m$(echo $nWalletAddress | jq -r .address)\033[0m"
			echo -e "PrivateKey: \033[33m$(echo $nWalletAddress | jq -r .private_key)\033[0m"
			break;;
		*   ) echo "Only 'yes' or 'no' please";;
	esac
done
        
sleep 2

if [ ! $nAddress ]; then
	read -p "Input your 0x Address: " nAddress
fi

echo -e "\e[1m\e[32m5. Checking your Memory & CPU... \e[0m" && sleep 1
nunet available --pretty

sleep 2

if [ ! $nMem ]; then
	read -p "Input Amount of Memory for NuNet: " nMem
fi

if [ ! $nCPU ]; then
	read -p "Input Amount of CPU for NuNet: " nCPU
fi

sleep 1
echo -e "\e[1m\e[32m6. Deleting trash... \e[0m" && sleep 1
rm nunet-dms-latest.deb

echo -e "\e[1m\e[32m7. Onboarding on NuNet... \e[0m" && sleep 1
nunet onboard -m $nMem -c $nCPU -n nunet-test -a $nAddress

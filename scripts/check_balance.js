const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  if (!deployer) {
    console.log("No se pudo obtener el signer del deployer. Verifica tu PRIVATE_KEY en .env.");
    return;
  }
  
  const address = await deployer.getAddress();
  console.log(`Billetera configurada: ${address}`);
  
  const provider = hre.ethers.provider;
  const balance = await provider.getBalance(address);
  console.log(`Balance en la red (${hre.network.name}): ${hre.ethers.formatEther(balance)} ETH`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

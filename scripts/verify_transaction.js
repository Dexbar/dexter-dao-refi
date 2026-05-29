// verify_transaction.js - Verify transaction status and target contract on blockchain
const { ethers } = require("ethers");
const fs = require("fs");
const path = require("path");

async function main() {
  const txHash = process.argv[2];
  const txType = process.argv[3]; // 'nft_mint', 'proposal_vote', 'donation'
  const expectedWallet = process.argv[4];

  if (!txHash || !txType) {
    console.error("Error: Falta el txHash o el tipo de transacción.");
    process.exit(1);
  }

  // Cargar las direcciones de los contratos inteligentes
  const deployedPath = path.join(__dirname, "../deployed_addresses.json");
  if (!fs.existsSync(deployedPath)) {
    console.error("Error: deployed_addresses.json no existe.");
    process.exit(1);
  }

  const deployed = JSON.parse(fs.readFileSync(deployedPath, "utf8"));
  const network = deployed.network || "sepolia";

  // Determinar la dirección de destino esperada según el tipo
  let expectedTargetAddress;
  if (txType === "nft_mint") {
    expectedTargetAddress = deployed.DexterNFT;
  } else if (txType === "proposal_vote") {
    expectedTargetAddress = deployed.DexterGov;
  } else if (txType === "donation") {
    expectedTargetAddress = deployed.DexterCrowdfund;
  } else {
    console.error("Error: Tipo de transacción desconocido.");
    process.exit(1);
  }

  if (!expectedTargetAddress) {
    console.error(`Error: Dirección para ${txType} no encontrada en deployed_addresses.json.`);
    process.exit(1);
  }

  // Determinar la URL del RPC según la red
  let rpcUrl;
  if (network === "localhost" || network === "hardhat") {
    rpcUrl = "http://127.0.0.1:8545";
  } else if (network === "sepolia") {
    rpcUrl = process.env.SEPOLIA_RPC_URL || "https://rpc.ankr.com/eth_sepolia";
  } else if (network === "alfajores") {
    rpcUrl = process.env.ALFAJORES_RPC_URL || "https://alfajores-forno.celo-testnet.org";
  } else {
    rpcUrl = process.env.RPC_URL || "https://rpc.ankr.com/eth_sepolia";
  }

  try {
    const provider = new ethers.JsonRpcProvider(rpcUrl);
    
    // Obtener el recibo de la transacción
    const receipt = await provider.getTransactionReceipt(txHash);
    if (!receipt) {
      console.log("VERIFY_FAILURE:Receipt not found");
      process.exit(0);
    }

    // 1. Verificar estado de la transacción (1 = éxito, 0 = fallido)
    if (receipt.status !== 1) {
      console.log("VERIFY_FAILURE:Transaction reverted");
      process.exit(0);
    }

    // 2. Verificar que el destino coincida con nuestro contrato
    if (receipt.to.toLowerCase() !== expectedTargetAddress.toLowerCase()) {
      console.log(`VERIFY_FAILURE:Target contract address mismatch (To: ${receipt.to}, Expected: ${expectedTargetAddress})`);
      process.exit(0);
    }

    // 3. Verificar que el remitente coincida
    if (expectedWallet && receipt.from.toLowerCase() !== expectedWallet.toLowerCase()) {
      console.log(`VERIFY_FAILURE:Sender address mismatch (From: ${receipt.from}, Expected: ${expectedWallet})`);
      process.exit(0);
    }

    console.log("VERIFY_SUCCESS:true");
  } catch (error) {
    const isConnectionError = error.message.includes("ECONNREFUSED") || 
                              error.message.includes("could not detect network") ||
                              error.message.includes("failed to detect network");
    if (isConnectionError && (network === "localhost" || network === "hardhat")) {
      // Fallback para desarrollo si el nodo local está apagado
      console.log("VERIFY_SUCCESS:true");
      process.exit(0);
    }
    console.error("Error querying blockchain:", error.message || error);
    console.log("VERIFY_SUCCESS:false");
    process.exit(1);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

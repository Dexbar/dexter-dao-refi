// scripts/blockchain_listener.js - Active daemon to listen to blockchain events and synchronize db.json
const { ethers } = require("ethers");
const fs = require("fs");
const path = require("path");

const DB_PATH = path.join(__dirname, "../db.json");

// Helper to load/save db.json safely
function readDB() {
  if (!fs.existsSync(DB_PATH)) return {};
  try {
    return JSON.parse(fs.readFileSync(DB_PATH, "utf8"));
  } catch (e) {
    console.error("Error reading DB:", e);
    return {};
  }
}

function writeDB(db) {
  try {
    fs.writeFileSync(DB_PATH, JSON.stringify(db, null, 2), "utf8");
  } catch (e) {
    console.error("Error writing DB:", e);
  }
}

function addLogToHistorial(perfil, message) {
  if (!perfil.Historial) perfil.Historial = [];
  perfil.Historial.push(message);
}

async function main() {
  console.log("📡 Starting Blockchain Event Listener Daemon...");

  const deployedPath = path.join(__dirname, "../deployed_addresses.json");
  if (!fs.existsSync(deployedPath)) {
    console.error("Error: deployed_addresses.json no existe. Saliendo.");
    process.exit(1);
  }

  const deployed = JSON.parse(fs.readFileSync(deployedPath, "utf8"));
  const network = deployed.network || "sepolia";

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

  console.log(`🔌 Connecting to RPC Provider: ${rpcUrl} (Network: ${network})`);
  const provider = new ethers.JsonRpcProvider(rpcUrl);

  // ABIs mínimas para los eventos
  const nftAbi = [
    "event NFTMinted(address indexed buyer, uint256 indexed tokenId, uint256 indexed nftType)"
  ];
  const govAbi = [
    "event Voted(address indexed voter, uint256 indexed proposalId, uint256 weight)"
  ];
  const crowdfundAbi = [
    "event DonationReceived(uint256 indexed campaignId, address indexed donor, uint256 amount)"
  ];

  const nftContract = new ethers.Contract(deployed.DexterNFT, nftAbi, provider);
  const govContract = new ethers.Contract(deployed.DexterGov, govAbi, provider);
  const crowdfundContract = new ethers.Contract(deployed.DexterCrowdfund, crowdfundAbi, provider);

  const nftKeys = ["ajolote", "luna", "quetzal", "androide", "supremo", "chinampero", "guardian"];
  const nftPrices = {
    ajolote: 200,
    luna: 500,
    quetzal: 1000,
    androide: 2000,
    supremo: 5000,
    chinampero: 1500,
    guardian: 3000
  };

  // Helper to handle NFTMinted
  async function handleNFTMinted(buyer, tokenId, nftType, txHash) {
    const wallet = buyer.toLowerCase();
    const typeId = Number(nftType);
    const key = nftKeys[typeId] || "unknown";
    const price = nftPrices[key] || 0;

    console.log(`👾 [NFTMinted] Buyer: ${wallet}, TokenId: ${tokenId}, Type: ${key}, Tx: ${txHash}`);

    const db = readDB();
    if (!db[wallet]) {
      db[wallet] = { Balance: 0, Inventario: [], Historial: [] };
    }

    const perfil = db[wallet];

    // Evitar duplicados
    const yaRegistrado = perfil.Historial.some(log => log.includes(txHash));
    if (yaRegistrado) {
      console.log(`   └─ Already registered, skipping.`);
      return;
    }

    // Actualizar perfil
    if (perfil.Balance >= price) {
      perfil.Balance -= price;
    } else {
      perfil.Balance = 0;
    }

    if (!perfil.Inventario) perfil.Inventario = [];
    perfil.Inventario.push(key);

    const now = new Date();
    const timeStr = `${String(now.getDate()).padStart(2, '0')}/${String(now.getMonth() + 1).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`;
    addLogToHistorial(perfil, `[${timeStr}] -${price} TK Compra en Blockchain de ${key} (Tx: ${txHash})`);

    db[wallet] = perfil;
    writeDB(db);
    console.log(`   └─ Sincronización exitosa.`);
  }

  // Helper to handle Voted
  async function handleVoted(voter, proposalId, weight, txHash) {
    const wallet = voter.toLowerCase();
    const propId = Number(proposalId);

    console.log(`🗳️ [Voted] Voter: ${wallet}, ProposalId: ${propId}, Tx: ${txHash}`);

    const db = readDB();
    if (!db[wallet]) {
      db[wallet] = { Balance: 0, Inventario: [], Historial: [] };
    }

    const perfil = db[wallet];

    // Evitar duplicados
    const yaRegistrado = perfil.Historial.some(log => log.includes(txHash));
    if (yaRegistrado) {
      console.log(`   └─ Already registered, skipping.`);
      return;
    }

    let propuesta = `Propuesta #${propId}`;
    if (propId === 0) {
      perfil.VotosRobot = (perfil.VotosRobot || 0) + 1;
      propuesta = "🤖 NUEVO NFT: ROBOT";
    } else if (propId === 1) {
      perfil.VotosPago = (perfil.VotosPago || 0) + 1;
      propuesta = "💰 SUBIR PAGO";
    }

    const now = new Date();
    const timeStr = `${String(now.getDate()).padStart(2, '0')}/${String(now.getMonth() + 1).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`;
    addLogToHistorial(perfil, `[${timeStr}] Voto firmado en Blockchain: ${propuesta} (Tx: ${txHash})`);

    db[wallet] = perfil;
    writeDB(db);
    console.log(`   └─ Sincronización exitosa.`);
  }

  // Helper to handle DonationReceived
  async function handleDonationReceived(campaignId, donor, amount, txHash) {
    const wallet = donor.toLowerCase();
    const campId = Number(campaignId);
    const amountEth = ethers.formatEther(amount); // O DXT según el formato

    console.log(`🎁 [DonationReceived] Donor: ${wallet}, Campaign: ${campId}, Amount: ${amountEth}, Tx: ${txHash}`);

    const db = readDB();
    if (!db[wallet]) {
      db[wallet] = { Balance: 0, Inventario: [], Historial: [] };
    }

    const perfil = db[wallet];

    // Evitar duplicados
    const yaRegistrado = perfil.Historial.some(log => log.includes(txHash));
    if (yaRegistrado) {
      console.log(`   └─ Already registered, skipping.`);
      return;
    }

    const now = new Date();
    const timeStr = `${String(now.getDate()).padStart(2, '0')}/${String(now.getMonth() + 1).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`;
    addLogToHistorial(perfil, `[${timeStr}] Donación firmada en Blockchain: ${amountEth} DXT a la Campaña #${campId} (Tx: ${txHash})`);

    db[wallet] = perfil;
    writeDB(db);
    console.log(`   └─ Sincronización exitosa.`);
  }

  // Poll new blocks and query logs to avoid event filter exceptions
  let lastBlock = -1;

  provider.on("block", async (blockNumber) => {
    if (blockNumber <= lastBlock) return;
    lastBlock = blockNumber;

    try {
      // 1. NFT Minted logs
      const nftLogs = await provider.getLogs({
        address: deployed.DexterNFT,
        fromBlock: blockNumber,
        toBlock: blockNumber
      });
      for (const log of nftLogs) {
        try {
          const parsed = nftContract.interface.parseLog(log);
          if (parsed && parsed.name === "NFTMinted") {
            const [buyer, tokenId, nftType] = parsed.args;
            await handleNFTMinted(buyer, tokenId, nftType, log.transactionHash);
          }
        } catch (err) {
          // Ignore unrelated logs
        }
      }

      // 2. Voted logs
      const govLogs = await provider.getLogs({
        address: deployed.DexterGov,
        fromBlock: blockNumber,
        toBlock: blockNumber
      });
      for (const log of govLogs) {
        try {
          const parsed = govContract.interface.parseLog(log);
          if (parsed && parsed.name === "Voted") {
            const [voter, proposalId, weight] = parsed.args;
            await handleVoted(voter, proposalId, weight, log.transactionHash);
          }
        } catch (err) {
          // Ignore unrelated logs
        }
      }

      // 3. DonationReceived logs
      const crowdfundLogs = await provider.getLogs({
        address: deployed.DexterCrowdfund,
        fromBlock: blockNumber,
        toBlock: blockNumber
      });
      for (const log of crowdfundLogs) {
        try {
          const parsed = crowdfundContract.interface.parseLog(log);
          if (parsed && parsed.name === "DonationReceived") {
            const [campaignId, donor, amount] = parsed.args;
            await handleDonationReceived(campaignId, donor, amount, log.transactionHash);
          }
        } catch (err) {
          // Ignore unrelated logs
        }
      }
    } catch (error) {
      console.error(`Error syncing events for block ${blockNumber}:`, error.message || error);
    }
  });

  console.log("🚀 Block logs polling active and running in background.");
  
  // Keep the process alive
  setInterval(() => {}, 1000);
}

main().catch(err => {
  console.error("Critical failure in listener daemon:", err);
  process.exit(1);
});

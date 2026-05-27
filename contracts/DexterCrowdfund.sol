// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract DexterCrowdfund is Ownable {
    IERC20 public dxtToken;

    struct Campaign {
        uint256 id;
        string title;
        string description;
        address payable recipient;
        uint256 goal;          // en DXT (con 18 decimales)
        uint256 amountRaised;  // en DXT (con 18 decimales)
        uint256 ethGoal;       // en Wei (opcional, 0 si no aplica)
        uint256 ethRaised;     // en Wei
        bool completed;
        bool fundsWithdrawn;
    }

    Campaign[] public campaigns;

    event CampaignCreated(
        uint256 indexed campaignId,
        string title,
        address recipient,
        uint256 goal,
        uint256 ethGoal
    );
    event DonationReceived(
        uint256 indexed campaignId,
        address indexed donor,
        uint256 amount,
        bool isDxt
    );
    event FundsWithdrawn(
        uint256 indexed campaignId,
        address recipient,
        uint256 dxtAmount,
        uint256 ethAmount
    );

    constructor(address _dxtTokenAddress) Ownable(msg.sender) {
        dxtToken = IERC20(_dxtTokenAddress);

        // Campañas iniciales de demostración
        _createCampaign(
            unicode"Rescate de Ajolotes en Xochimilco",
            unicode"Fondo para restaurar el hábitat natural de los ajolotes y crear canales limpios en Xochimilco.",
            payable(msg.sender),
            5000 * 10**18, // 5000 DXT
            1 * 10**18    // 1 ETH
        );
        _createCampaign(
            unicode"Educación Web3 y Programación Go",
            unicode"Fondo comunitario para capacitar a jóvenes de escuelas locales en tecnologías blockchain y Go.",
            payable(msg.sender),
            2500 * 10**18, // 2500 DXT
            0.5 * 10**18  // 0.5 ETH
        );
    }

    function _createCampaign(
        string memory _title,
        string memory _description,
        address payable _recipient,
        uint256 _goal,
        uint256 _ethGoal
    ) internal {
        uint256 campaignId = campaigns.length;
        campaigns.push(Campaign({
            id: campaignId,
            title: _title,
            description: _description,
            recipient: _recipient,
            goal: _goal,
            amountRaised: 0,
            ethGoal: _ethGoal,
            ethRaised: 0,
            completed: false,
            fundsWithdrawn: false
        }));
        emit CampaignCreated(campaignId, _title, _recipient, _goal, _ethGoal);
    }

    // Crear una nueva campaña (Cualquier usuario puede proponer/crear una campaña)
    function createCampaign(
        string calldata _title,
        string calldata _description,
        address payable _recipient,
        uint256 _goal,
        uint256 _ethGoal
    ) external {
        require(_recipient != address(0), "Direccion invalida");
        _createCampaign(_title, _description, _recipient, _goal, _ethGoal);
    }

    // Donar DXT tokens a una campaña
    function donateDXT(uint256 _campaignId, uint256 _amount) external {
        require(_campaignId < campaigns.length, "Campana no existe");
        Campaign storage camp = campaigns[_campaignId];
        require(!camp.fundsWithdrawn, "Los fondos ya fueron retirados");

        // Transferir del donante al contrato (debe haber aprobado el contrato antes)
        require(dxtToken.transferFrom(msg.sender, address(this), _amount), "Transferencia DXT fallida");

        camp.amountRaised += _amount;
        if (camp.amountRaised >= camp.goal && (camp.ethGoal == 0 || camp.ethRaised >= camp.ethGoal)) {
            camp.completed = true;
        }

        emit DonationReceived(_campaignId, msg.sender, _amount, true);
    }

    // Donar ETH a una campaña
    function donateETH(uint256 _campaignId) external payable {
        require(_campaignId < campaigns.length, "Campana no existe");
        Campaign storage camp = campaigns[_campaignId];
        require(!camp.fundsWithdrawn, "Los fondos ya fueron retirados");
        require(msg.value > 0, "Monto debe ser mayor a 0");

        camp.ethRaised += msg.value;
        if (camp.ethRaised >= camp.ethGoal && camp.amountRaised >= camp.goal) {
            camp.completed = true;
        }

        emit DonationReceived(_campaignId, msg.sender, msg.value, false);
    }

    // Retirar fondos (Solo el receptor de la campaña puede retirar los fondos)
    function withdrawFunds(uint256 _campaignId) external {
        require(_campaignId < campaigns.length, "Campana no existe");
        Campaign storage camp = campaigns[_campaignId];
        require(msg.sender == camp.recipient, "No eres el beneficiario");
        require(!camp.fundsWithdrawn, "Fondos ya retirados");
        
        camp.fundsWithdrawn = true;
        uint256 dxtToWithdraw = camp.amountRaised;
        uint256 ethToWithdraw = camp.ethRaised;

        // Enviar DXT
        if (dxtToWithdraw > 0) {
            require(dxtToken.transfer(camp.recipient, dxtToWithdraw), "Fallo al enviar DXT");
        }

        // Enviar ETH
        if (ethToWithdraw > 0) {
            (bool success, ) = camp.recipient.call{value: ethToWithdraw}("");
            require(success, "Fallo al enviar ETH");
        }

        emit FundsWithdrawn(_campaignId, camp.recipient, dxtToWithdraw, ethToWithdraw);
    }

    function getCampaignsCount() external view returns (uint256) {
        return campaigns.length;
    }

    function getCampaign(uint256 _campaignId) external view returns (
        uint256 id,
        string memory title,
        string memory description,
        address recipient,
        uint256 goal,
        uint256 amountRaised,
        uint256 ethGoal,
        uint256 ethRaised,
        bool completed,
        bool fundsWithdrawn
    ) {
        require(_campaignId < campaigns.length, "Campana no existe");
        Campaign memory camp = campaigns[_campaignId];
        return (
            camp.id,
            camp.title,
            camp.description,
            camp.recipient,
            camp.goal,
            camp.amountRaised,
            camp.ethGoal,
            camp.ethRaised,
            camp.completed,
            camp.fundsWithdrawn
        );
    }
}

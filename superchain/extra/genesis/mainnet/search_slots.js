const fs = require('fs');

const EIP1967_PROXY_ADMIN = "0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103";
const PROXY_ADMIN_STORAGE_VALUE = "0x0000000000000000000000004200000000000000000000000000000000000018";

let proxyAdminOccurrences = 0;
let additionalSlots = 0;
let controlledContracts = [];

const genesisData = JSON.parse(fs.readFileSync('./blaineup.json', 'utf8'));
for (let allocation in genesisData.alloc) {
    const alloc = genesisData.alloc;
    const addressAllocation = alloc[allocation];
    if (addressAllocation.storage) {
        const storage = addressAllocation.storage;
        for (let slot in storage) {
            const value = storage[slot];
            if (slot == EIP1967_PROXY_ADMIN) {
                proxyAdminOccurrences++;
                if (value == PROXY_ADMIN_STORAGE_VALUE) {
                    controlledContracts.push(allocation);
                } else {
                    console.error(`Something isn't right, storage slot value incorrect for ${allocation}`);
                }
            } else {
                if (value == PROXY_ADMIN_STORAGE_VALUE) {
                    throw new Error("L2 ProxyAdmin address used for unexpected slot");
                }
            }
        }
    }
}

console.log("Total eip1967.proxy.admin: ", proxyAdminOccurrences);
console.log("Number of contracts controlled by eip1967.proxy.admin: ", controlledContracts.length);
console.log("Additional slots: ", additionalSlots);

const timestamp = Date.now();
const outputFile = `proxy_admin_controlled_contracts_${timestamp}.json`;
const jsonData = JSON.stringify({ data: controlledContracts }, null, 2);

try {
 fs.writeFileSync(outputFile, jsonData, 'utf8');
 console.log(`Data written to ${outputFile}`);
} catch (err) {
 console.error('Error writing file:', err);
}

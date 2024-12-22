package io.optimism.embed;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import io.optimism.superchain.AddressList;
import org.hyperledger.besu.datatypes.Address;

import java.io.IOException;
import java.util.Optional;

/**
 * The type AddressList deserializer.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class AddressListDeserializer extends JsonDeserializer<AddressList> {
    @Override
    public AddressList deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
        JsonNode node = p.getCodec().readTree(p);
        AddressList addressList = new AddressList();
        addressList.setAddressManager(Address.fromHexStringStrict(node.get("AddressManager").asText()));
        addressList.setGuardian(Address.fromHexStringStrict(node.get("Guardian").asText()));
        addressList.setL1CrossDomainMessengerProxy(Address.fromHexStringStrict(node.get("L1CrossDomainMessengerProxy").asText()));
        addressList.setL1ERC721BridgeProxy(Address.fromHexStringStrict(node.get("L1ERC721BridgeProxy").asText()));
        addressList.setL1StandardBridgeProxy(Address.fromHexStringStrict(node.get("L1StandardBridgeProxy").asText()));
        addressList.setOptimismMintableERC20FactoryProxy(Address.fromHexStringStrict(node.get("OptimismMintableERC20FactoryProxy").asText()));
        addressList.setOptimismPortalProxy(Address.fromHexStringStrict(node.get("OptimismPortalProxy").asText()));
        addressList.setSystemConfigProxy(Address.fromHexStringStrict(node.get("SystemConfigProxy").asText()));
        addressList.setSystemConfigOwner(Address.fromHexStringStrict(node.get("SystemConfigOwner").asText()));
        addressList.setProxyAdmin(Address.fromHexStringStrict(node.get("ProxyAdmin").asText()));
        addressList.setProxyAdminOwner(Address.fromHexStringStrict(node.get("ProxyAdminOwner").asText()));
        if (node.get("Challenger") != null) {
            addressList.setChallenger(Optional.of(Address.fromHexStringStrict(node.get("Challenger").asText())));
        }
        if (node.get("L2OutputOracleProxy") != null) {
            addressList.setL2OutputOracleProxy(Optional.of(Address.fromHexStringStrict(node.get("L2OutputOracleProxy").asText())));
        }
        if (node.get("AnchorStateRegistryProxy") != null) {
            addressList.setAnchorStateRegistryProxy(Optional.of(Address.fromHexStringStrict(node.get("AnchorStateRegistryProxy").asText())));
        }
        if (node.get("DelayedWETHProxy") != null) {
            addressList.setDelayedWETHProxy(Optional.of(Address.fromHexStringStrict(node.get("DelayedWETHProxy").asText())));
        }
        if (node.get("DisputeGameFactoryProxy") != null) {
            addressList.setDisputeGameFactoryProxy(Optional.of(Address.fromHexStringStrict(node.get("DisputeGameFactoryProxy").asText())));
        }
        if (node.get("FaultDisputeGame") != null) {
            addressList.setFaultDisputeGame(Optional.of(Address.fromHexStringStrict(node.get("FaultDisputeGame").asText())));
        }
        if (node.get("MIPS") != null) {
            addressList.setMips(Optional.of(Address.fromHexStringStrict(node.get("MIPS").asText())));
        }
        if (node.get("PermissionedDisputeGame") != null) {
            addressList.setPermissionedDisputeGame(Optional.of(Address.fromHexStringStrict(node.get("PermissionedDisputeGame").asText())));
        }
        if (node.get("PreimageOracle") != null) {
            addressList.setPreimageOracle(Optional.of(Address.fromHexStringStrict(node.get("PreimageOracle").asText())));
        }
        return addressList;
    }
}

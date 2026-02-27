use crate::msg::{
    AddressOfResponse, ContractInfoResponse, GetParentIdResponse,
    IsContractResponse,
    MemeArtNftInfoResponse,
    OriginalAdminResponse,
};
use crate::state::{MINTING_FEES_INFO, ORIGINAL_ADMIN_INFO};
use crate::Cw721MetadataContract;
use cosmwasm_std::{Deps, StdError, StdResult};

// version info for migration info
pub const CONTRACT_NAME: &str = "crates.io:memeart";
pub const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

pub fn contract_info(deps: Deps) -> StdResult<ContractInfoResponse> {
    let minting_fees = MINTING_FEES_INFO.load(deps.storage)?;

    let contract_info_response = ContractInfoResponse {
        contract_name: CONTRACT_NAME.to_string(),
        contract_version: CONTRACT_VERSION.to_string(),
        native_denom: minting_fees.native_denom,
        native_decimals: minting_fees.native_decimals,
        token_cap: minting_fees.token_cap,
        base_mint_fee: minting_fees.base_mint_fee,
        burn_percentage: minting_fees.burn_percentage,
        original_percentage: minting_fees.original_percentage,
        parent_percentage: minting_fees.parent_percentage,
        dao_address: minting_fees.dao_address,
        funds_address: minting_fees.funds_address,
    };
    Ok(contract_info_response)
}

pub fn original_admin(deps: Deps) -> StdResult<OriginalAdminResponse> {
    let original_admin = ORIGINAL_ADMIN_INFO.load(deps.storage)?;

    let original_admin_response = OriginalAdminResponse {
        original_admin_address1: original_admin.original_admin_address1,
        original_admin_address2: original_admin.original_admin_address2,
        original_admin_address3: original_admin.original_admin_address3,
    };
    Ok(original_admin_response)
}

pub fn is_contract(
    contract: Cw721MetadataContract,
    deps: Deps,
    token_id: String,
) -> StdResult<IsContractResponse> {
    let token = contract.tokens.load(deps.storage, &token_id)?;

    if let Some(addr) = token.extension.contract_address {
        Ok(IsContractResponse {
            contract_address: addr,
        })
    } else {
        Err(StdError::NotFound {
            kind: "No contract address".to_string(),
        })
    }
}

// like owner_of
// but returns owner
// and contract address (or none)
pub fn address_of(
    contract: Cw721MetadataContract,
    deps: Deps,
    token_id: String,
) -> StdResult<AddressOfResponse> {
    let token = contract.tokens.load(deps.storage, &token_id)?;
    Ok(AddressOfResponse {
        owner: token.owner.to_string(),
        contract_address: token.extension.contract_address,
        validator_address: token.extension.validator_operator_address,
    })
}

// looks up the actual token
// so throws an error if it doesn't exist
pub fn get_parent_id(
    contract: Cw721MetadataContract,
    deps: Deps,
    token_id: String,
) -> StdResult<GetParentIdResponse> {
    let token = contract.tokens.load(deps.storage, &token_id)?;

    match token.extension.parent_token_id {
        Some(pti) => {
            // attempt to load parent
            // else error
            let _parent_token = contract.tokens.load(deps.storage, &pti)?;

            Ok(GetParentIdResponse {
                parent_token_id: pti,
            })
        }
        None => Err(StdError::NotFound {
            kind: "Parent not found".to_string(),
        }),
    }
}

pub fn get_parent_nft_info(
    contract: Cw721MetadataContract,
    deps: Deps,
    token_id: String,
) -> StdResult<MemeArtNftInfoResponse> {
    let token = contract.tokens.load(deps.storage, &token_id)?;

    match token.extension.parent_token_id {
        Some(pti) => {
            // attempt to load parent
            let parent_token = contract.tokens.load(deps.storage, &pti)?;

            Ok(MemeArtNftInfoResponse {
                token_uri: parent_token.token_uri,
                extension: parent_token.extension,
            })
        }
        None => Err(StdError::NotFound {
            kind: "Parent not found".to_string(),
        }),
    }
}

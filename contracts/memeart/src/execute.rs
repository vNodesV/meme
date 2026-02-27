use crate::error::ContractError;
use cosmwasm_std::{ensure_eq, Binary, DepsMut, Env, MessageInfo, Response, StdResult, Addr};
use cw2::set_contract_version;
use cw721::Cw721ReceiveMsg;
use cw721_base::state::TokenInfo;
use cw_utils::{must_pay, one_coin};

use std::convert::TryInto;

use crate::msg::{
    InstantiateMsg, Metadata, MintMsg, MintingFeesResponse, UpdateMetadataMsg,
    UpdateMintingFeesMsg, UpdateOriginImageMsg, UpdateOriginAdminMsg, OriginalAdminResponse
};

use crate::state::{MINTING_FEES_INFO, ORIGINAL_ADMIN_INFO};
use crate::utils::{
    get_mint_response, get_number_of_owned_tokens,
};
use crate::Cw721MetadataContract;

// version info for migration info
pub const CONTRACT_NAME: &str = "crates.io:memeart";
pub const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

pub fn execute_instantiate(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let minting_fees = MintingFeesResponse {
        native_denom: msg.native_denom,
        native_decimals: msg.native_decimals,
        token_cap: msg.token_cap,
        base_mint_fee: msg.base_mint_fee,
        burn_percentage: msg.burn_percentage,
        original_percentage: msg.original_percentage,
        parent_percentage: msg.parent_percentage,
        dao_address: msg.dao_address,
        funds_address: msg.funds_address,
    };
    MINTING_FEES_INFO.save(deps.storage, &minting_fees)?;
    let admin_address = deps.api.addr_validate(&msg.admin_address)?;
    contract.minter.save(deps.storage, &admin_address)?;
    Ok(Response::default())
}

// update minting fees
pub fn update_minting_fees(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: UpdateMintingFeesMsg,
) -> Result<Response, ContractError> {
    let address_trying_to_update = info.sender;

    // look up contract admin
    let current_admin_address = contract.minter(deps.as_ref())?.minter;

    // check it's the admin of the contract updating
    ensure_eq!(
        current_admin_address,
        address_trying_to_update,
        ContractError::Unauthorized {}
    );

    // get current fees
    let minting_fees_info = MINTING_FEES_INFO.load(deps.storage)?;

    let minting_fees = MintingFeesResponse {
        // these two can't be updated
        native_denom: minting_fees_info.native_denom,
        native_decimals: minting_fees_info.native_decimals,
        // these can
        token_cap: msg.token_cap,
        base_mint_fee: msg.base_mint_fee,
        burn_percentage: msg.burn_percentage,
        original_percentage: msg.original_percentage,
        parent_percentage: msg.parent_percentage,
        dao_address: msg.dao_address,
        funds_address: msg.funds_address,
    };

    // update
    MINTING_FEES_INFO.save(deps.storage, &minting_fees)?;

    let res = Response::new().add_attribute("action", "update_contract_minting_fees");
    Ok(res)
}

// this actually updates the ADMIN address, but under the hood it is
// called minter by the contract.
// On the query side we actually just proxy to the existing Minter query
pub fn set_admin_address(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    admin_address: String,
) -> Result<Response, ContractError> {
    let address_trying_to_update = info.sender;
    let current_admin_address = contract.minter(deps.as_ref())?.minter;

    // check it's the admin of the contract updating
    ensure_eq!(
        current_admin_address,
        address_trying_to_update,
        ContractError::Unauthorized {}
    );

    // validate
    let validated_addr = deps.api.addr_validate(&admin_address)?;

    // update
    contract.minter.save(deps.storage, &validated_addr)?;

    let res = Response::new()
        .add_attribute("action", "update_contract_admin_address")
        .add_attribute("new_admin_address", validated_addr);
    Ok(res)
}

// boy oh boy this needs a refactor
pub fn mint(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: MintMsg,
) -> Result<Response, ContractError> {
    // any address can mint
    // sender of the execute
    let address_trying_to_mint = info.sender.clone();

    // can only mint NFTs belonging to yourself
    ensure_eq!(
        msg.owner,
        address_trying_to_mint,
        ContractError::Unauthorized {}
    );

    // get minting fees and minter (i.e. admin)
    let minting_fees = MINTING_FEES_INFO.load(deps.storage)?;
    let minter = contract.minter(deps.as_ref())?.minter;
    let admin_address = deps.api.addr_validate(&minter)?;

    // check if trying to mint too many
    // who can need more than 20?
    let default_limit: usize = 4000000;
    let number_of_tokens_owned = get_number_of_owned_tokens(
        &contract,
        &deps,
        address_trying_to_mint.clone(),
        default_limit,
    )?;

    // error out if we exceed configured cap or we already
    // have the default max
    match minting_fees.token_cap {
        Some(tc) => {
            if number_of_tokens_owned >= tc.try_into().unwrap() {
                return Err(ContractError::TokenCapExceeded {});
            }
        }
        None => {
            if number_of_tokens_owned >= default_limit {
                return Err(ContractError::TokenCapExceeded {});
            }
        }
    }

    struct SurchargeAddr {
        parent_address: Option<Addr>,
        original_address: Option<Addr>,
    }

    let mut new_extension = msg.extension;

    let mut surcharge:SurchargeAddr = SurchargeAddr {
        parent_address: None,
        original_address: None,
    };

    let mut original_image:bool = false;

    match new_extension.parent_token_id {
        Some(ref _pti) => {
            if !_pti.is_empty() && !_pti.eq(&msg.token_id) {
                // find parent token if not empty
                let _parent_token = contract.tokens.load(deps.storage, &_pti.to_string())?;
                surcharge.parent_address = Some(deps.api.addr_validate(&_parent_token.owner.to_string())?);

                // find original token from parent
                match _parent_token.extension.original_token_id {
                    Some(ref _oti) => {
                        new_extension.original_token_id = Some(_oti.to_string());
                        let _original_token = contract.tokens.load(deps.storage, &_oti.to_string())?;
                        surcharge.original_address = Some(deps.api.addr_validate(&_original_token.owner.to_string())?);
                        match _original_token.extension.original_image {
                            Some(is_original) => {
                                if is_original == true {
                                    original_image = true;
                                }
                            }
                            None => {}
                        }
                    }
                    None => {
                        new_extension.original_token_id = Some(_pti.to_string());
                        let _original_token = contract.tokens.load(deps.storage, &_pti.to_string())?;
                        surcharge.original_address = Some(deps.api.addr_validate(&_original_token.owner.to_string())?);
                        match _original_token.extension.original_image {
                            Some(is_original) => {
                                if is_original == true {
                                    original_image = true;
                                }
                            }
                            None => {}
                        }
                    }
                }
            } else {
                new_extension.parent_token_id = Some("".to_string());
            }
        }
        None => {}
    }

    // validate owner addr
    let owner_address = deps.api.addr_validate(&msg.owner)?;

    // work out what fees are owed
    let fee = minting_fees.base_mint_fee;
    let coin = one_coin(&info)?;
    // error out if this fee isn't covered in the msg
    if fee.is_some() {
        must_pay(&info, &minting_fees.native_denom)?;

        // ensure atomicity
        
        if let Some(fee_amount) = fee {
            if coin.amount < fee_amount {
                return Err(ContractError::InsufficientFunds {});
            }
        }
    };

    let username = &msg.token_id.to_lowercase();

    // create the token
    // this will fail if token_id (i.e. username)
    // is already claimed
    let token = TokenInfo {
        owner: owner_address,
        approvals: vec![],
        token_uri: msg.token_uri,
        extension: new_extension,
    };
    contract
        .tokens
        .update(deps.storage, username, |old| match old {
            Some(_) => Err(ContractError::Claimed {}),
            None => Ok(token),
        })?;

    contract.increment_tokens(deps.storage)?;

    // if there is a fee, add a bank msg to send to the admin_address
    let res = get_mint_response(
        admin_address,
        address_trying_to_mint,
        surcharge.parent_address,
        surcharge.original_address,
        minting_fees.native_denom,
        Some(coin.amount),
        minting_fees.parent_percentage,
        minting_fees.original_percentage,
        msg.token_id,
        original_image,
        minting_fees.dao_address,
        minting_fees.funds_address
    );
    Ok(res)
}

// updates the metadata on an NFT
// only accessible to the NFT owner
// note that the parent_token_id field
// is immutable and cannot be updated
pub fn update_metadata(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: UpdateMetadataMsg,
) -> Result<Response, ContractError> {
    let address_trying_to_update = info.sender.clone();
    let token_id = msg.token_id.clone();
    let username_nft = contract.tokens.load(deps.storage, &token_id)?;

    let username_owner = username_nft.owner.clone();

    // this is immutable
    let existing_parent_id = username_nft.extension.parent_token_id.clone();

    // check it's the owner of the NFT updating meta
    ensure_eq!(
        username_owner,
        address_trying_to_update,
        ContractError::Unauthorized {}
    );

    // arrrrre you ready to rrrrrumb-
    // rrredefine some metadata?
    contract
        .tokens
        .update(deps.storage, &token_id, |token| -> StdResult<_> {
            match token {
                Some(mut nft) => {
                    nft.extension = msg.metadata;
                    nft.extension.parent_token_id = existing_parent_id;
                    Ok(nft)
                }
                None => Ok(username_nft),
            }
        })?;

    Ok(Response::new()
        .add_attribute("action", "update_metadata")
        .add_attribute("owner", info.sender)
        .add_attribute("token_id", token_id))
}

// updates the original admin address
// only accessible to the admin address
pub fn update_origin_admin(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: UpdateOriginAdminMsg,
) -> Result<Response, ContractError> {
    let address_trying_to_update = info.sender.clone();
    let current_admin_address = contract.minter(deps.as_ref())?.minter;


    // check it's the admin of the contract updating
    ensure_eq!(
        current_admin_address,
        address_trying_to_update,
        ContractError::Unauthorized {}
    );

    let origin_admin = OriginalAdminResponse {
        original_admin_address1: msg.original_admin_address1,
        original_admin_address2: msg.original_admin_address2,
        original_admin_address3: msg.original_admin_address3,
    };

    // update
    ORIGINAL_ADMIN_INFO.save(deps.storage, &origin_admin)?;

    let res = Response::new().add_attribute("action", "update_origin_admin");
    Ok(res)
}

// updates the original_image on an NFT
// only accessible to the admin address
pub fn update_origin_image(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: UpdateOriginImageMsg,
) -> Result<Response, ContractError> {
    let address_trying_to_update = Some(info.sender.clone());
    //let current_admin_address = contract.minter(deps.as_ref())?.minter;

    let admins = ORIGINAL_ADMIN_INFO.load(deps.storage)?;

    let admin_address1 = Some(deps.api.addr_validate(&admins.original_admin_address1)?);
    let admin_address2 = Some(deps.api.addr_validate(&admins.original_admin_address2)?);
    let admin_address3 = Some(deps.api.addr_validate(&admins.original_admin_address3)?);

    let mut allow_admin1 = true;
    let mut allow_admin2 = true;
    let mut allow_admin3 = true;

    // check it's the admin of the origin image update
    if admin_address1 != address_trying_to_update {
        allow_admin1 = false;
    }
    if admin_address2 != address_trying_to_update {
        allow_admin2 = false;
    }
    if admin_address3 != address_trying_to_update {
        allow_admin3 = false;
    }
    

    if !allow_admin1 && !allow_admin2 && !allow_admin3 {
        return Err(ContractError::Unauthorized {});
    }


    let token_id = msg.token_id.clone();
    let username_nft = contract.tokens.load(deps.storage, &token_id)?;

    // arrrrre you ready to rrrrrumb-
    // rrredefine some metadata?
    contract
        .tokens
        .update(deps.storage, &token_id, |token| -> StdResult<_> {
            match token {
                Some(mut nft) => {
                    nft.extension.original_image = msg.original_image;
                    Ok(nft)
                }
                None => Ok(username_nft),
            }
        })?;

    Ok(Response::new()
        .add_attribute("action", "update_metadata")
        .add_attribute("owner", info.sender)
        .add_attribute("token_id", token_id))
}

// this function clears metadata
// for situations like transfer and send
// to enable web of trust stuff
// and make sure stale meta doesn't persist after send/transfer
pub fn clear_metadata(deps: DepsMut, token_id: String) -> Result<(), ContractError> {
    let contract = Cw721MetadataContract::default();
    let username_nft = contract.tokens.load(deps.storage, &token_id)?;
    contract
        .tokens
        .update(deps.storage, &token_id, |token| -> StdResult<_> {
            match token {
                Some(mut nft) => {
                    nft.extension = Metadata {
                        ..Metadata::default()
                    };
                    Ok(nft)
                }
                None => Ok(username_nft),
            }
        })?;
    Ok(())
}

pub fn transfer_nft(
    contract: Cw721MetadataContract,
    mut deps: DepsMut,
    env: Env,
    info: MessageInfo,
    recipient: String,
    token_id: String,
) -> Result<Response, ContractError> {
    // check permissions before proceeding
    let token = contract.tokens.load(deps.storage, &token_id)?;
    contract.check_can_send(deps.as_ref(), &env, &info, &token)?;

    // blank meta before xfer
    clear_metadata(deps.branch(), token_id.to_string())?;

    contract._transfer_nft(deps, &env, &info, &recipient, &token_id)?;

    Ok(Response::new()
        .add_attribute("action", "transfer_nft")
        .add_attribute("sender", info.sender)
        .add_attribute("recipient", recipient)
        .add_attribute("token_id", token_id))
}

pub fn send_nft(
    contract: Cw721MetadataContract,
    mut deps: DepsMut,
    env: Env,
    info: MessageInfo,
    receiving_contract: String,
    token_id: String,
    msg: Binary,
) -> Result<Response, ContractError> {
    // check permissions before proceeding
    let token = contract.tokens.load(deps.storage, &token_id)?;
    contract.check_can_send(deps.as_ref(), &env, &info, &token)?;

    // blank meta before send
    clear_metadata(deps.branch(), token_id.to_string())?;

    // Transfer token
    contract._transfer_nft(deps, &env, &info, &receiving_contract, &token_id)?;

    let send = Cw721ReceiveMsg {
        sender: info.sender.to_string(),
        token_id: token_id.clone(),
        msg,
    };

    // Send message
    Ok(Response::new()
        .add_message(send.into_cosmos_msg(receiving_contract.clone())?)
        .add_attribute("action", "send_nft")
        .add_attribute("sender", info.sender)
        .add_attribute("recipient", receiving_contract)
        .add_attribute("token_id", token_id))
}

pub fn burn(
    contract: Cw721MetadataContract,
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    token_id: String,
) -> Result<Response, ContractError> {
    let token = contract.tokens.load(deps.storage, &token_id)?;
    contract.check_can_send(deps.as_ref(), &env, &info, &token)?;

    contract.tokens.remove(deps.storage, &token_id)?;
    contract.decrement_tokens(deps.storage)?;

    Ok(Response::new()
        .add_attribute("action", "burn")
        .add_attribute("sender", info.sender)
        .add_attribute("token_id", token_id))
}

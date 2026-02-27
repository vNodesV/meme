use cosmwasm_std::{
    coins, Addr, BankMsg, CosmosMsg, Decimal, DepsMut, Order, Response, StdResult,
    Uint128,
};
use crate::Cw721MetadataContract;

pub fn get_number_of_owned_tokens(
    contract: &Cw721MetadataContract,
    deps: &DepsMut,
    address: Addr,
    default_limit: usize,
) -> StdResult<usize> {
    let owned_tokens: Vec<String> = contract
        .tokens
        .idx
        .owner
        .prefix(address)
        .keys(deps.storage, None, None, Order::Ascending)
        .take(default_limit) // set default big limit
        .collect::<StdResult<Vec<_>>>()?;

    let number_of_tokens_owned = owned_tokens.len();
    Ok(number_of_tokens_owned)
}

pub fn get_mint_response(
    admin_address: Addr,
    mint_message_sender: Addr,
    _parent_address: Option<Addr>,
    _original_address: Option<Addr>,
    native_denom: String,
    fee: Option<Uint128>,
    _parent_percentage: Option<u64>,
    _original_percentage: Option<u64>,
    token_id: String,
    original_image: bool,
    dao_address: Option<String>,
    funds_address: Option<String>,
) -> Response {
    match fee {

        Some(fee) => {

            let fee_to_original = fee * Decimal::percent(_original_percentage.unwrap());
            let fee_to_parent = fee * Decimal::percent(_parent_percentage.unwrap());

            let mut msgs: Vec<CosmosMsg> = Vec::new();
            let mut fee_to_funds = fee;
            let mut fee_to_dao = 0;

            fee_to_funds -= fee_to_original;
            fee_to_funds -= fee_to_parent;

            let mut original_to_dao = true;
            let mut parent_to_dao = true;

            if original_image == true {
                match _original_address {
                    Some(oa) => {
                        original_to_dao = false;
                        msgs.push(BankMsg::Send {
                            to_address: oa.to_string(),
                            amount: coins(fee_to_original.u128(), native_denom.clone()),
                        }
                        .into());
                    }
                    None => {}
                };
            }
            match _parent_address {
                Some(pa) => {
                    parent_to_dao = false;
                    msgs.push(BankMsg::Send {
                        to_address: pa.to_string(),
                        amount: coins(fee_to_parent.u128(), native_denom.clone()),
                    }
                    .into());
                }
                None => {}
            };
            if original_to_dao == true {
                fee_to_dao += fee_to_original.u128();
            }
            if parent_to_dao == true {
                fee_to_dao += fee_to_parent.u128();
            }

            if fee_to_dao > 0 {
                match dao_address {
                    Some(address) => {
                        msgs.push(BankMsg::Send {
                            to_address: address.to_string(),
                            amount: coins(fee_to_dao, native_denom.clone()),
                        }
                        .into());
                    }
                    None => {
                        fee_to_funds += Uint128::from(fee_to_dao);
                    }
                }
            }

            match funds_address {
                Some(address) => {
                    msgs.push(BankMsg::Send {
                        to_address: address.to_string(),
                        amount: coins(fee_to_funds.u128(), native_denom.clone()),
                    }
                    .into());
                }
                None => {
                    msgs.push(BankMsg::Send {
                        to_address: admin_address.to_string(),
                        amount: coins(fee_to_funds.u128(), native_denom.clone()),
                    }
                    .into());
                }
            }

            Response::new()
                .add_attribute("action", "mint")
                .add_attribute("minter", mint_message_sender)
                .add_attribute("token_id", token_id)
                .add_messages(msgs)
        }
        None => Response::new()
            .add_attribute("action", "mint")
            .add_attribute("minter", mint_message_sender)
            .add_attribute("token_id", token_id),
    }
}
// Copyright 2023-2024, Offchain Labs, Inc.
// For licensing, see https://github.com/OffchainLabs/stylus-sdk-rs/blob/main/licenses/COPYRIGHT.md

use proc_macro::TokenStream;
use quote::ToTokens;
use syn::{parse_macro_input, parse_quote};

use crate::imports::stylus_sdk::abi::AbiType;

/// Implementation of the [`#[derive(AbiType)]`][crate::AbiType] macro.
pub fn derive_abi_type(input: TokenStream) -> TokenStream {
    let item = parse_macro_input!(input as syn::ItemStruct);
    impl_abi_type(&item).into_token_stream().into()
}

/// Implement [`stylus_sdk::abi::AbiType`] for the given struct.
///
/// The name is used for the ABI name to match the
/// [`SolType::SOL_NAME`][alloy_sol_types::SolType::SOL_NAME] generated by the
/// [`sol!`][alloy_sol_types::sol] macro.
fn impl_abi_type(item: &syn::ItemStruct) -> syn::ItemImpl {
    let name = &item.ident;
    let name_str = name.to_string();
    let (impl_generics, ty_generics, where_clause) = item.generics.split_for_impl();

    parse_quote! {
        impl #impl_generics #AbiType for #name #ty_generics #where_clause {
            type SolType = Self;

            const ABI: stylus_sdk::abi::ConstString = stylus_sdk::abi::ConstString::new(#name_str);
        }
    }
}

#[cfg(test)]
mod tests {
    use syn::parse_quote;

    use super::impl_abi_type;
    use crate::utils::testing::assert_ast_eq;

    #[test]
    fn test_impl_abi_type() {
        assert_ast_eq(
            impl_abi_type(&parse_quote! {
                struct Foo<T>
                where T: Bar {
                    a: bool,
                    b: String,
                    t: T,
                }
            }),
            parse_quote! {
                impl<T> stylus_sdk::abi::AbiType for Foo<T>
                where T: Bar {
                    type SolType = Self;

                    const ABI: stylus_sdk::abi::ConstString = stylus_sdk::abi::ConstString::new("Foo");
                }
            },
        )
    }
}

extern crate alloc;
extern crate core;
extern crate wee_alloc;

use anyhow::{anyhow, Result};

use k8s_openapi::api::core::v1 as apicore;

use alloc::vec::Vec;
use std::mem::MaybeUninit;
use std::env;

/// Logs a message to the console using [`_log`].
fn log(message: String) {
    unsafe {
        let (ptr, len) = string_to_ptr(&message);
        _log(ptr, len);
    }
}

#[link(wasm_import_module = "env")]
extern "C" {
    /// WebAssembly import which prints a string (linear memory offset,
    /// byteCount) to the console.
    ///
    /// Note: This is not an ownership transfer: Rust still owns the pointer
    /// and ensures it isn't deallocated during this call.
    #[link_name = "log"]
    fn _log(ptr: u32, size: u32);
}

// /// WebAssembly export that accepts a string (linear memory offset, byteCount)
// /// and calls [`greet`].
// ///
// /// Note: The input parameters were returned by [`allocate`]. This is not an
// /// ownership transfer, so the inputs can be reused after this call.
// #[cfg_attr(all(target_arch = "wasm32"), export_name = "greet")]
// #[no_mangle]
// pub unsafe extern "C" fn _greet(ptr: u32, len: u32) {
//     greet(&ptr_to_string(ptr, len));
// }

/// Returns a pointer and size pair for the given string in a way compatible
/// with WebAssembly numeric types.
///
/// Note: This doesn't change the ownership of the String. To intentionally
/// leak it, use [`std::mem::forget`] on the input after calling this.
unsafe fn string_to_ptr(s: &String) -> (u32, u32) {
    return (s.as_ptr() as u32, s.len() as u32);
}

/// Set the global allocator to the WebAssembly optimized one.
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

/// WebAssembly export that allocates a pointer (linear memory offset) that can
/// be used for a string.
///
/// This is an ownership transfer, which means the caller must call
/// [`deallocate`] when finished.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "allocate")]
#[no_mangle]
pub extern "C" fn _allocate(size: u32) -> *mut u8 {
    allocate(size as usize)
}

/// Allocates size bytes and leaks the pointer where they start.
fn allocate(size: usize) -> *mut u8 {
    // Allocate the amount of bytes needed.
    let vec: Vec<MaybeUninit<u8>> = Vec::with_capacity(size);

    // into_raw leaks the memory to the caller.
    Box::into_raw(vec.into_boxed_slice()) as *mut u8
}

/// WebAssembly export that deallocates a pointer of the given size (linear
/// memory offset, byteCount) allocated by [`allocate`].
#[cfg_attr(all(target_arch = "wasm32"), export_name = "deallocate")]
#[no_mangle]
pub unsafe extern "C" fn _deallocate(ptr: u32, size: u32) {
    deallocate(ptr as *mut u8, size as usize);
}

/// Retakes the pointer which allows its memory to be freed.
unsafe fn deallocate(ptr: *mut u8, size: usize) {
    let _ = Vec::from_raw_parts(ptr, 0, size);
}

fn eval() {
    // let args: Vec<_> = env::args().collect();
    // let object_to_test = &args[1];
    let object_to_test = r#"{"apiVersion":"v1","kind":"Pod","metadata":{"name":"nginx","labels":{"app":"nginx"}},"spec":{"containers":[{"name":"nginx","image":"nginx","securityContext":{"privileged":false}}]}}"#;

    match serde_json::from_str::<apicore::Pod>(object_to_test) {
        Ok(pod) => {
            if let Some(pod_spec) = &pod.spec {
                return match validate_pod(pod_spec) {
                    Ok(_) => println!("true"),
                    Err(_) => println!("false"),
                };
            };
            // If there is not pod spec, just accept it. There is no data to be
            // validated.
            println!("true")
        }
        Err(_) => println!("false"),
    }
}

/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and calls [`eval`].
///
/// Note: The input parameters were returned by [`allocate`]. This is not an
/// ownership transfer, so the inputs can be reused after this call.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "eval")]
#[no_mangle]
pub unsafe extern "C" fn _eval() {
    eval();
}

fn validate_pod(pod: &apicore::PodSpec) -> Result<bool> {
    for container in &pod.containers {
        let container_valid = validate_container(container);
        if !container_valid {
            return Err(anyhow!("Privileged container is not allowed"));
        }
    }
    if let Some(init_containers) = &pod.init_containers {
        for container in init_containers {
            let container_valid = validate_container(container);
            if !container_valid {
                return Err(anyhow!("Privileged init container is not allowed"));
            }
        }
    }
    if let Some(ephemeral_containers) = &pod.ephemeral_containers {
        for container in ephemeral_containers {
            let container_valid = validate_ephemeral_container(container);
            if !container_valid {
                return Err(anyhow!("Privileged ephemeral container is not allowed"));
            }
        }
    }
    Ok(true)
}

fn validate_ephemeral_container(container: &apicore::EphemeralContainer) -> bool {
    if let Some(security_context) = &container.security_context {
        return !security_context.privileged.unwrap_or(false);
    }
    true
}

fn validate_container(container: &apicore::Container) -> bool {
    if let Some(security_context) = &container.security_context {
        return !security_context.privileged.unwrap_or(false);
    }
    true
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn accept_pod_when_all_ephemeral_containers_are_not_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            ephemeral_containers: Some(vec![
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_ok(),
            "Pod with no privileged ephemeral container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_all_ephemeral_container_is_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            ephemeral_containers: Some(vec![
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_err(),
            "Pod with all privileged ephemeral container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_one_ephemeral_container_is_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            ephemeral_containers: Some(vec![
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
                apicore::EphemeralContainer {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::EphemeralContainer::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(result.is_err(),
            "Pod with only a single privileged ephemeral container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_pod_when_init_containers_are_not_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            init_containers: Some(vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_ok(),
            "Pod with no privileged init container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_one_init_container_is_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            init_containers: Some(vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_err(),
            "Pod with only a single privileged init container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_all_init_containers_are_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            init_containers: Some(vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ]),
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_err(),
            "Pod with all privileged init containers should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn accecpt_pod_when_containers_are_not_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            containers: vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ],
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_ok(),
            "Pod with no privileged container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_one_container_is_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            containers: vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(false),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ],
            ..apicore::PodSpec::default()
        });

        assert!(
            result.is_err(),
            "Pod with only a single privileged container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_pod_when_all_containers_are_privileged_test() -> Result<()> {
        let result = validate_pod(&apicore::PodSpec {
            containers: vec![
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
                apicore::Container {
                    security_context: Some(apicore::SecurityContext {
                        privileged: Some(true),
                        ..apicore::SecurityContext::default()
                    }),
                    ..apicore::Container::default()
                },
            ],
            ..apicore::PodSpec::default()
        });
        assert!(
            result.is_err(),
            "Pod with all privileged containers should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_container_is_not_privileged_test() -> Result<()> {
        assert_eq!(
            validate_container(&apicore::Container {
                security_context: Some(apicore::SecurityContext {
                    privileged: Some(false),
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::Container::default()
            }),
            true,
            "Non privileged container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_container_with_no_security_context() -> Result<()> {
        assert_eq!(
            validate_container(&apicore::Container {
                ..apicore::Container::default()
            }),
            true,
            "Non privileged container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_privileged_container_test() -> Result<()> {
        assert_eq!(
            validate_container(&apicore::Container {
                security_context: Some(apicore::SecurityContext {
                    privileged: Some(true),
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::Container::default()
            }),
            false,
            "Privileged container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_privileged_container_when_privileged_is_none_test() -> Result<()> {
        assert_eq!(
            validate_container(&apicore::Container {
                security_context: Some(apicore::SecurityContext {
                    privileged: None,
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::Container::default()
            }),
            true,
            "Privileged container should be accepted by the validator when there is no 'privileged' configuration. The default behaviour is disable privileged containers"
        );
        Ok(())
    }

    #[test]
    fn accept_ephemeral_container_is_not_privileged_test() -> Result<()> {
        assert_eq!(
            validate_ephemeral_container(&apicore::EphemeralContainer {
                security_context: Some(apicore::SecurityContext {
                    privileged: Some(false),
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::EphemeralContainer::default()
            }),
            true,
            "Non privileged container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_ephemeral_container_with_no_security_context() -> Result<()> {
        assert_eq!(
            validate_ephemeral_container(&apicore::EphemeralContainer {
                ..apicore::EphemeralContainer::default()
            }),
            true,
            "Non privileged container should be accepted by the validator"
        );
        Ok(())
    }

    #[test]
    fn reject_privileged_ephemeral_container_test() -> Result<()> {
        assert_eq!(
            validate_ephemeral_container(&apicore::EphemeralContainer {
                security_context: Some(apicore::SecurityContext {
                    privileged: Some(true),
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::EphemeralContainer::default()
            }),
            false,
            "Privileged container should be rejected by the validator"
        );
        Ok(())
    }

    #[test]
    fn accept_privileged_ephemeral_container_when_privileged_is_none_test() -> Result<()> {
        assert_eq!(
            validate_ephemeral_container(&apicore::EphemeralContainer {
                security_context: Some(apicore::SecurityContext {
                    privileged: None,
                    ..apicore::SecurityContext::default()
                }),
                ..apicore::EphemeralContainer::default()
            }),
            true,
            "Privileged container should be accepted by the validator when there is no 'privileged' configuration. The default behaviour is disable privileged containers"
        );
        Ok(())
    }
}

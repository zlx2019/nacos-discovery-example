fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("cargo:rerun-if-changed=./build.rs");
    println!("cargo:rerun-if-changed=./proto/Envelope.proto");
    tonic_prost_build::configure()
        .compile_protos(&["./proto/Envelope.proto"], &["./proto"])
        .unwrap();
    Ok(())
}
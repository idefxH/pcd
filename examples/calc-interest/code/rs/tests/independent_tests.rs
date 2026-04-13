// generated from spec: calc-interest.spec.md sha256:8279a6f935e0a8c1f7e3caa355a553dcb6470960bb2d87b6b4cda99caa48f941
// independent_tests — integration tests for calc-interest (Rust)
// Each test function corresponds to a named EXAMPLE in the spec.
// Tests invoke the compiled binary via std::process::Command.
//
// Run with: cargo test --release --test independent_tests
//
// Confidence mapping (see TRANSLATION_REPORT.md):
//   test_typical_calculation       → EXAMPLE: typical_calculation       (High)
//   test_zero_rate_rejected        → EXAMPLE: zero_rate_rejected         (High)
//   test_zero_principal_rejected   → EXAMPLE: zero_principal_rejected    (High)
//   test_zero_periods_rejected     → EXAMPLE: zero_periods_rejected      (High)
//   test_non_numeric_input_rejected→ EXAMPLE: non_numeric_input_rejected (High)
//   test_version_output              → EXAMPLE: version_output              (High)

use std::io::Write;
use std::process::{Command, Stdio};

/// Helper: run the binary with the given stdin bytes, return (stdout, stderr, exit_code).
fn run(input: &[u8]) -> (String, String, i32) {
    // The binary is expected at target/release/calc-interest relative to the
    // workspace root (where `cargo test` is invoked).
    let mut child = Command::new(env!("CARGO_BIN_EXE_calc-interest"))
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("failed to spawn calc-interest binary");

    child
        .stdin
        .take()
        .expect("stdin not available")
        .write_all(input)
        .expect("failed to write to stdin");

    let output = child.wait_with_output().expect("failed to wait for child");

    let stdout = String::from_utf8_lossy(&output.stdout).into_owned();
    let stderr = String::from_utf8_lossy(&output.stderr).into_owned();
    let code = output.status.code().unwrap_or(-1);
    (stdout, stderr, code)
}

/// EXAMPLE: typical_calculation
/// principal=10000.00, rate=0.0350, periods=12
/// Expected: INTEREST: 4200.00 / TOTAL: 14200.00 / exit 0
#[test]
fn test_typical_calculation() {
    let (stdout, stderr, code) = run(b"10000.00\n0.0350\n12\n");
    assert_eq!(code, 0, "exit code should be 0; stderr: {}", stderr);
    assert!(stderr.is_empty(), "stderr should be empty on success; got: {}", stderr);
    let lines: Vec<&str> = stdout.lines().collect();
    assert_eq!(lines.len(), 2, "stdout should have exactly 2 lines; got: {:?}", lines);
    assert_eq!(lines[0], "INTEREST: 4200.00");
    assert_eq!(lines[1], "TOTAL:    14200.00");
}

/// EXAMPLE: zero_rate_rejected
/// rate=0.0000 → stderr contains "invalid rate", exit 2
#[test]
fn test_zero_rate_rejected() {
    let (stdout, stderr, code) = run(b"10000.00\n0.0000\n12\n");
    assert_eq!(code, 2, "exit code should be 2; stderr: {}", stderr);
    assert!(
        stderr.contains("invalid rate"),
        "stderr should contain 'invalid rate'; got: {}",
        stderr
    );
    assert!(stdout.is_empty(), "stdout should be empty on error; got: {}", stdout);
}

/// EXAMPLE: zero_principal_rejected
/// principal=0.00 → stderr contains "invalid principal", exit 2
#[test]
fn test_zero_principal_rejected() {
    let (stdout, stderr, code) = run(b"0.00\n0.0350\n12\n");
    assert_eq!(code, 2, "exit code should be 2; stderr: {}", stderr);
    assert!(
        stderr.contains("invalid principal"),
        "stderr should contain 'invalid principal'; got: {}",
        stderr
    );
    assert!(stdout.is_empty(), "stdout should be empty on error; got: {}", stdout);
}

/// EXAMPLE: zero_periods_rejected
/// periods=0 → stderr contains "invalid periods", exit 2
#[test]
fn test_zero_periods_rejected() {
    let (stdout, stderr, code) = run(b"10000.00\n0.0350\n0\n");
    assert_eq!(code, 2, "exit code should be 2; stderr: {}", stderr);
    assert!(
        stderr.contains("invalid periods"),
        "stderr should contain 'invalid periods'; got: {}",
        stderr
    );
    assert!(stdout.is_empty(), "stdout should be empty on error; got: {}", stdout);
}

/// EXAMPLE: non_numeric_input_rejected
/// principal="abc" → stderr contains error message, exit 1
#[test]
fn test_non_numeric_input_rejected() {
    let (stdout, stderr, code) = run(b"abc\n0.0350\n12\n");
    assert_eq!(code, 1, "exit code should be 1; stderr: {}", stderr);
    assert!(
        !stderr.is_empty(),
        "stderr should contain an error message; got empty"
    );
    assert!(stdout.is_empty(), "stdout should be empty on error; got: {}", stdout);
}

/// EXAMPLE: version_output
/// invoked with "version" → stdout matches "calc-interest 0.2.0 spec:{64-hex-chars}", exit 0
#[test]
fn test_version_output() {
    let mut child = std::process::Command::new(env!("CARGO_BIN_EXE_calc-interest"))
        .arg("version")
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .expect("failed to spawn calc-interest binary");

    let output = child.wait_with_output().expect("failed to wait for child");
    let stdout = String::from_utf8_lossy(&output.stdout).into_owned();
    let stderr = String::from_utf8_lossy(&output.stderr).into_owned();
    let code = output.status.code().unwrap_or(-1);

    assert_eq!(code, 0, "exit code should be 0; stderr: {}", stderr);
    assert!(stderr.is_empty(), "stderr should be empty; got: {}", stderr);
    let line = stdout.trim();
    assert!(
        line.starts_with("calc-interest "),
        "output must start with 'calc-interest '; got: {}",
        line
    );
    assert!(
        line.contains(" spec:"),
        "output must contain ' spec:'; got: {}",
        line
    );
    // Verify exact version and 64-hex-char sha256
    assert!(
        line == "calc-interest 0.2.0 spec:8279a6f935e0a8c1f7e3caa355a553dcb6470960bb2d87b6b4cda99caa48f941",
        "version output mismatch; got: {}",
        line
    );
}

import { assertEquals } from "@std/assert";

export type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

export function ok(): Result<void>;
export function ok<T>(value: T): Result<T>;
export function ok<T>(value?: T): Result<T | void> {
  return { ok: true, value: value };
}

export function err<E>(error: E): Result<never, E> {
  return { ok: false, error: error };
}

export function unwrap<T>(result: Result<T>): T {
  if (!result.ok) {
    throw new Error(`Result is not ok. Error: ${result.error}`);
  }
  return result.value;
}

export function resultFromFn<T>(
  fn: () => T,
): Result<T> {
  try {
    return ok(fn());
  } catch (e) {
    return err(e as Error);
  }
}

export function assertResultEquals<T>(
  actual: Result<T>,
  expected: T,
  msg?: string,
) {
  if (!actual.ok) {
    throw new AssertionError(
      `Result is not ok. Error: ${actual.error}`,
    );
  }

  assertEquals(actual.value, expected, msg);
}

export class AssertionError extends Error {
  /** Constructs a new instance.
   *
   * @param message The error message.
   * @param options Additional options. This argument is still unstable. It may change in the future release.
   */
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "AssertionError";
  }
}

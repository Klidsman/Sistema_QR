import { expect, Page } from '@playwright/test';

export const correoAdministrador = process.env.E2E_CORREO ?? 'admin@colegio.pe';
export const claveAdministrador = process.env.E2E_CLAVE ?? 'clave-segura-2026';
export const seccionDemostracion = process.env.E2E_SECCION ?? '3.° A (2026)';

export async function ingresarComoAdministrador(page: Page): Promise<void> {
  await page.goto('/ingreso');
  await page.getByLabel('Correo').fill(correoAdministrador);
  await page.getByLabel('Contraseña').fill(claveAdministrador);
  await page.getByRole('button', { name: 'Ingresar' }).click();
  await expect(page).toHaveURL(/\/alumnos$/);
}

// identificadorUnico evita choques entre ejecuciones repetidas de las pruebas.
export function identificadorUnico(): string {
  const azar = Math.floor(Math.random() * 10_000).toString().padStart(4, '0');
  return '9' + Date.now().toString().slice(-6) + azar;
}

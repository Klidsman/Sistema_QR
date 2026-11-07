import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Porteros', () => {
  test('el administrador crea una cuenta y el portero solo ingresa a portería', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const nombre = `Portero ${sufijo}`;
    const correo = `portero${sufijo}@colegio.pe`;
    const clave = 'clave-portero-2026';

    await page.goto('/porteros');
    await page.getByLabel('Nombre completo *').fill(nombre);
    await page.getByLabel('Correo *').fill(correo);
    await page.getByLabel('Contraseña *').fill(clave);
    await page.getByRole('button', { name: 'Crear portero' }).click();

    await expect(page.getByText('La cuenta de portero se creó correctamente.')).toBeVisible();
    const fila = page.locator('tr', { hasText: correo });
    await expect(fila).toBeVisible();
    await expect(fila.getByText('Activo')).toBeVisible();

    await page.getByRole('button', { name: 'Salir' }).click();
    await expect(page).toHaveURL(/\/ingreso$/);

    await page.getByLabel('Correo').fill(correo);
    await page.getByLabel('Contraseña').fill(clave);
    await page.getByRole('button', { name: 'Ingresar' }).click();
    await expect(page).toHaveURL(/\/porteria$/);
    await expect(page.getByRole('heading', { name: 'Registro de ingreso' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Porteros', exact: true })).toHaveCount(0);
  });

  test('rechaza un correo repetido', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const correo = `portero.repetido${sufijo}@colegio.pe`;

    for (const intento of [1, 2]) {
      await page.goto('/porteros');
      await page.getByLabel('Nombre completo *').fill(`Portero ${intento} ${sufijo}`);
      await page.getByLabel('Correo *').fill(correo);
      await page.getByLabel('Contraseña *').fill('clave-portero-2026');
      await page.getByRole('button', { name: 'Crear portero' }).click();
    }

    await expect(page.getByText('Ya existe una cuenta con ese correo.')).toBeVisible();
  });
});

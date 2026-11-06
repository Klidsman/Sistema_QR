import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Docentes', () => {
  test('el administrador crea, restablece la clave y desactiva una cuenta', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const nombre = `Docente ${sufijo}`;
    const correo = `docente${sufijo}@colegio.pe`;

    await page.goto('/docentes');
    await page.getByLabel('Nombre completo *').fill(nombre);
    await page.getByLabel('Correo *').fill(correo);
    await page.getByLabel('Contraseña *').fill('clave-docente-2026');
    await page.getByRole('button', { name: 'Crear docente' }).click();

    await expect(page.getByText('La cuenta de docente se creó correctamente.')).toBeVisible();
    const fila = page.locator('tr', { hasText: correo });
    await expect(fila).toBeVisible();
    await expect(fila.getByText('Activo')).toBeVisible();

    await fila.getByLabel('Nueva clave').fill('otra-clave-2026');
    await fila.getByRole('button', { name: 'Restablecer' }).click();
    await expect(page.getByText('La contraseña se restableció correctamente.')).toBeVisible();

    page.once('dialog', (dialogo) => dialogo.accept());
    await page.locator('tr', { hasText: correo }).getByRole('button', { name: 'Desactivar' }).click();
    await expect(page.getByText('La cuenta de docente se desactivó correctamente.')).toBeVisible();
    await expect(page.locator('tr', { hasText: correo }).getByText('Inactivo')).toBeVisible();
  });

  test('rechaza un correo repetido', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const sufijo = Date.now().toString().slice(-5);
    const correo = `repetido${sufijo}@colegio.pe`;

    for (const intento of [1, 2]) {
      await page.goto('/docentes');
      await page.getByLabel('Nombre completo *').fill(`Docente ${intento} ${sufijo}`);
      await page.getByLabel('Correo *').fill(correo);
      await page.getByLabel('Contraseña *').fill('clave-docente-2026');
      await page.getByRole('button', { name: 'Crear docente' }).click();
    }

    await expect(page.getByText('Ya existe una cuenta con ese correo.')).toBeVisible();
  });
});

import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Página central', () => {
  test('presenta el colegio y los accesos por rol', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('heading', { name: /Asistencia clara/ })).toBeVisible();
    await expect(page.getByText('María Parado de Bellido').first()).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Cuatro pasos, de la puerta al reporte' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Cómo funciona' })).toBeVisible();

    for (const rol of ['administrador', 'direccion', 'porteria', 'docente']) {
      const acceso = page.locator(`#acceso-${rol}`);
      await expect(acceso).toBeVisible();
      await expect(acceso).toHaveAttribute('href', '/ingreso');
    }

    await page.locator('#acceso-porteria').click();
    await expect(page).toHaveURL(/\/ingreso$/);
    await expect(page.getByRole('heading', { name: 'Ingreso al sistema' })).toBeVisible();
  });

  test('con sesión iniciada ofrece ir al panel del rol', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/');

    const panel = page.getByRole('link', { name: 'Ir a mi panel' }).first();
    await expect(panel).toHaveAttribute('href', '/alumnos');
  });
});

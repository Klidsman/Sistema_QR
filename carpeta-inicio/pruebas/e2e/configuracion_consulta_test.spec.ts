import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

test.describe('Historia 5: consulta de configuración histórica', () => {
  test('muestra la configuración vigente e informa cuando no hay año', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/configuracion/consulta');

    await expect(page.getByRole('heading', { name: 'Consulta histórica' })).toBeVisible();
    await expect(page.getByText(/Año 2026/)).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Turno de las secciones' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Horarios por turno' })).toBeVisible();

    await page.getByLabel('Fecha').fill('2020-01-01');
    await page.getByRole('button', { name: 'Consultar' }).click();

    await expect(page.getByText('No hay configuración vigente para la fecha seleccionada.')).toBeVisible();
  });
});

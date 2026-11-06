import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function fechaEnDias(dias: number): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() + dias);
  return fecha.toISOString().slice(0, 10);
}

test.describe('Historia 4: umbral de faltas seguidas', () => {
  test('configura el umbral de faltas seguidas', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/configuracion/umbrales');

    await page.getByLabel('Tipo *').selectOption({ label: 'Faltas seguidas' });
    await page.getByLabel('Valor *').fill('2');
    await page.getByLabel('Vigencia desde *').fill(fechaEnDias(0));
    await page.getByRole('button', { name: 'Guardar umbral' }).click();

    await expect(page.getByText('El umbral de alerta se guardó correctamente.')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Faltas seguidas' })).toBeVisible();
    await expect(page.getByText('Vigente desde').first()).toBeVisible();
  });
});

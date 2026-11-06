import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function fechaEnDias(dias: number): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() + dias);
  return fecha.toISOString().slice(0, 10);
}

test.describe('Historia 4: umbrales de alerta', () => {
  test('configura tardanzas y programa un cambio futuro', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/configuracion/umbrales');

    await page.getByLabel('Tipo *').selectOption({ label: 'Tardanzas' });
    await page.getByLabel('Valor *').fill('3');
    await page.getByLabel('Vigencia desde *').fill(fechaEnDias(0));
    await page.getByRole('button', { name: 'Guardar umbral' }).click();

    await expect(page.getByText('El umbral de alerta se guardó correctamente.')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Tardanzas por mes' })).toBeVisible();

    await page.getByLabel('Tipo *').selectOption({ label: 'Tardanzas' });
    await page.getByLabel('Valor *').fill('4');
    await page.getByLabel('Vigencia desde *').fill(fechaEnDias(30));
    await page.getByRole('button', { name: 'Guardar umbral' }).click();

    await expect(page.getByText('El umbral de alerta se guardó correctamente.')).toBeVisible();
    await expect(page.getByRole('cell', { name: '4', exact: true })).toBeVisible();
  });
});

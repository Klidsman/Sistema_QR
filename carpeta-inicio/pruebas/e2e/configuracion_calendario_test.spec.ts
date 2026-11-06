import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function fechaEnDias(dias: number): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() + dias);
  return fecha.toISOString().slice(0, 10);
}

test.describe('Historia 3: feriados y suspensiones', () => {
  test('registra un feriado y confirma un solapamiento', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const motivo = `Feriado ${Date.now().toString().slice(-5)}`;
    const fecha = fechaEnDias(40);

    await page.goto('/configuracion/calendario');
    await page.getByLabel('Tipo *').selectOption({ label: 'Feriado' });
    await page.getByLabel('Motivo *').fill(motivo);
    await page.getByLabel('Fecha de inicio *').fill(fecha);
    await page.getByLabel('Fecha de fin *').fill(fecha);
    await page.getByRole('button', { name: 'Registrar excepción' }).click();
    await page.waitForLoadState('load');
    const confirmarPrimero = page.getByRole('button', { name: 'Confirmar y guardar' });
    if ((await confirmarPrimero.count()) > 0) {
      await confirmarPrimero.click();
    }

    await expect(page.getByText('La excepción de calendario se registró correctamente.')).toBeVisible();
    await expect(page.getByText(motivo)).toBeVisible();
    const filaFeriado = page.locator('tr', { hasText: motivo });
    await expect(filaFeriado.getByRole('cell', { name: 'Todo el colegio', exact: true })).toBeVisible();

    await page.getByLabel('Tipo *').selectOption({ label: 'Suspensión' });
    await page.getByLabel('Motivo *').fill(motivo + ' suspensión');
    await page.getByLabel('Fecha de inicio *').fill(fecha);
    await page.getByLabel('Fecha de fin *').fill(fecha);
    await page.getByRole('button', { name: 'Registrar excepción' }).click();

    await expect(page.getByText('Confirmar y guardar')).toBeVisible();
    await page.getByRole('button', { name: 'Confirmar y guardar' }).click();

    await expect(page.getByText('La excepción de calendario se registró correctamente.')).toBeVisible();
    await expect(page.getByText(motivo + ' suspensión')).toBeVisible();
  });
});

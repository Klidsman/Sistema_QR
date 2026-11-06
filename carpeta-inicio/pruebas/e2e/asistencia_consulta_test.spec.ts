import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function hoyISO(): string {
  return new Date().toISOString().slice(0, 10);
}

test.describe('Correcciones y plazo de asistencia', () => {
  test('configura el plazo y corrige un registro con motivo', async ({ page }) => {
    await ingresarComoAdministrador(page);

    await page.goto('/configuracion/asistencia');
    await page.getByLabel('Plazo de edición en horas *').fill('1');
    await page.getByRole('button', { name: 'Guardar plazo', exact: true }).click();
    await expect(page.getByText('El plazo de edición se guardó correctamente.')).toBeVisible();
    await expect(page.getByText('1 horas')).toBeVisible();

    await page.goto('/asistencia/consulta');
    await page.getByLabel('Sección').selectOption({ label: '1.° A (2026)' });
    await page.getByLabel('Fecha').fill(hoyISO());
    await page.getByRole('button', { name: 'Consultar' }).click();
    await expect(page.getByRole('heading', { name: 'Registros de la sección' })).toBeVisible();

    const correccion = page.locator('form[action="/asistencia/consulta/correccion"]').first();
    if ((await correccion.count()) > 0) {
      await correccion.locator('select[name="estado"]').selectOption('presente');
      await correccion.locator('input[name="motivo"]').fill('Corrección de prueba E2E');
      await correccion.getByRole('button', { name: 'Corregir' }).click();
      await expect(page.getByText('El registro se corrigió correctamente.')).toBeVisible();
    } else {
      await expect(page.getByText('No hay registros del docente para esa fecha.')).toBeVisible();
    }
  });
});

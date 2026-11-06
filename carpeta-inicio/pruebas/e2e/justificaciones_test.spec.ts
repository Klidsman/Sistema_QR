import { expect, test } from '@playwright/test';

import { crearAlumnoConQR, ingresarComoAdministrador } from './ayudas';

function diaLaborableReciente(): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() - 1);
  while (fecha.getDay() === 0 || fecha.getDay() === 6) {
    fecha.setDate(fecha.getDate() - 1);
  }
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Lima' }).format(fecha);
}

test.describe('Historia 1: registrar una justificación', () => {
  test('el administrador registra una justificación pendiente y no puede autoaprobarla', async ({ page }) => {
    await ingresarComoAdministrador(page);
    const alumno = await crearAlumnoConQR(page);
    const urlFicha = page.url().replace(/\?.*$/, '');
    const alumnoID = urlFicha.split('/').pop();

    await page.goto(`/justificaciones/nueva?alumno_id=${alumnoID}`);
    await page.getByLabel('Desde *').fill(diaLaborableReciente());
    await page.getByLabel('Hasta *').fill(diaLaborableReciente());
    await page.getByLabel('Motivo *').fill('Reposo médico presentado por la familia');
    await page.getByRole('button', { name: 'Registrar justificación' }).click();

    await expect(page.getByText('La justificación quedó pendiente de aprobación.')).toBeVisible();
    await expect(page.getByText(alumno.nombre, { exact: false }).first()).toBeVisible();
    await expect(page.getByText('Pendiente').first()).toBeVisible();

    await page.locator('article a[href^="/justificaciones/"]').first().click();
    await expect(page.getByRole('button', { name: 'Aprobar' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Anular justificación' })).toHaveCount(0);
  });
});

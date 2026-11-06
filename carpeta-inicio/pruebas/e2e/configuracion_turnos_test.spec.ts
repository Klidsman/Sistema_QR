import { expect, test } from '@playwright/test';

import { ingresarComoAdministrador } from './ayudas';

function fechaEnDias(dias: number): string {
  const fecha = new Date();
  fecha.setDate(fecha.getDate() + dias);
  return fecha.toISOString().slice(0, 10);
}

test.describe('Historia 2: horarios y tolerancias por turno', () => {
  test('guarda el horario de mañana y lo muestra vigente', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/configuracion/turnos');

    const formulario = page.locator('form[action="/configuracion/turnos/manana"]');
    await formulario.getByLabel('Hora de ingreso (Mañana) *').fill('07:30');
    await formulario.getByLabel('Tolerancia en minutos (Mañana) *').fill('10');
    await formulario.getByLabel('Hora límite de tardanza (Mañana) *').fill('08:00');
    await formulario.getByLabel('Hora de cierre (Mañana) *').fill('08:30');
    await formulario.getByLabel('Vigencia desde (Mañana) *').fill(fechaEnDias(0));
    await formulario.getByRole('button', { name: 'Guardar horario' }).click();

    await expect(page.getByText('El horario del turno se guardó correctamente.')).toBeVisible();
    await expect(page.getByRole('definition').filter({ hasText: '07:30' })).toBeVisible();
    await expect(page.getByRole('definition').filter({ hasText: '10 min' })).toBeVisible();
  });

  test('rechaza una hora de cierre anterior al límite de tardanza', async ({ page }) => {
    await ingresarComoAdministrador(page);
    await page.goto('/configuracion/turnos');

    const formulario = page.locator('form[action="/configuracion/turnos/tarde"]');
    await formulario.getByLabel('Hora de ingreso (Tarde) *').fill('12:45');
    await formulario.getByLabel('Tolerancia en minutos (Tarde) *').fill('10');
    await formulario.getByLabel('Hora límite de tardanza (Tarde) *').fill('13:15');
    await formulario.getByLabel('Hora de cierre (Tarde) *').fill('13:00');
    await formulario.getByLabel('Vigencia desde (Tarde) *').fill(fechaEnDias(0));
    await formulario.getByRole('button', { name: 'Guardar horario' }).click();

    await expect(page.getByText('La hora de cierre debe ser posterior a la hora límite de tardanza.')).toBeVisible();
  });
});

<?php

use Illuminate\Support\Facades\Schema;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Database\Migrations\Migration;

class CreateBetsTable extends Migration
{
    /**
     * Run the migrations.
     *
     * @return void
     */
    public function up()
    {
        Schema::create('bets', function (Blueprint $table) {
            $table->increments('id');
            $table->unsignedInteger('value');
            $table->boolean('winner')->default(false);
            $table->dateTime('created_at');
            $table->dateTime('deleted_at')->nullable();
            $table->unsignedInteger('lot_id');
            $table->unsignedInteger('user_id');

            $table->index('deleted_at');
            $table->index('lot_id');
            $table->index('user_id');

            $table->foreign('lot_id')->references('id')->on('lots');
            $table->foreign('user_id')->references('id')->on('users');
        });
    }

    /**
     * Reverse the migrations.
     *
     * @return void
     */
    public function down()
    {
        Schema::dropIfExists('bets');
    }
}
